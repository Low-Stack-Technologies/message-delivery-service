package auth

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"golang.org/x/crypto/ssh"
)

func NewMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := r.Header.Get("X-Client-Id")
			timestampStr := r.Header.Get("X-Timestamp")
			authHeader := r.Header.Get("Authorization")

			config.DebugLog("[DEBUG] Auth Attempt - ClientID: %s, Timestamp: %s, Auth: %s", clientID, timestampStr, authHeader)

			if clientID == "" || timestampStr == "" || authHeader == "" {
				config.DebugLog("[DEBUG] Auth Failed - Missing headers")
				http.Error(w, "Missing authentication headers", http.StatusUnauthorized)
				return
			}

			// 1. Verify Timestamp (Replay Protection)
			timestamp, err := time.Parse(time.RFC3339, timestampStr)
			if err != nil {
				config.DebugLog("[DEBUG] Auth Failed - Invalid timestamp format: %v", err)
				http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
				return
			}
			if time.Since(timestamp) > 5*time.Minute || time.Since(timestamp) < -5*time.Minute {
				config.DebugLog("[DEBUG] Auth Failed - Timestamp expired: diff=%v", time.Since(timestamp))
				http.Error(w, "Request timestamp expired or in the future", http.StatusUnauthorized)
				return
			}

			// 2. Find Service & Public Key
			cfg := config.Get()
			var service *config.ServiceConfig
			for _, s := range cfg.Services {
				if s.ID == clientID {
					service = &s
					break
				}
			}

			if service == nil {
				config.DebugLog("[DEBUG] Auth Failed - Unknown Client ID: %s", clientID)
				http.Error(w, "Unknown Client ID", http.StatusUnauthorized)
				return
			}

			pubKey, err := parsePublicKey(service.PublicKey)
			if err != nil {
				config.DebugLog("[DEBUG] Auth Failed - Public key parsing error: %v", err)
				http.Error(w, "Service public key is misconfigured", http.StatusInternalServerError)
				return
			}

			// 3. Construct Canonical Request
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for later handlers

			bodyHash := sha256.Sum256(bodyBytes)
			bodyHashHex := hex.EncodeToString(bodyHash[:])

			// Canonical = Method + "\n" + Path + "\n" + X-Timestamp + "\n" + SHA256(Body)
			canonical := r.Method + "\n" + r.URL.Path + "\n" + timestampStr + "\n" + bodyHashHex
			config.DebugLog("[DEBUG] Canonical Request:\n%s", canonical)

			// 4. Verify Signature
			if len(authHeader) < 10 || authHeader[:10] != "Signature " {
				config.DebugLog("[DEBUG] Auth Failed - Invalid Auth header format")
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			signatureB64 := authHeader[10:]
			signature, err := base64.StdEncoding.DecodeString(signatureB64)
			if err != nil {
				config.DebugLog("[DEBUG] Auth Failed - Signature decode error: %v", err)
				http.Error(w, "Invalid signature encoding", http.StatusUnauthorized)
				return
			}

			if !ed25519.Verify(pubKey, []byte(canonical), signature) {
				config.DebugLog("[DEBUG] Auth Failed - Ed25519 verification failed")
				http.Error(w, "Signature verification failed", http.StatusUnauthorized)
				return
			}

			config.DebugLog("[DEBUG] Auth Success - ClientID: %s", clientID)
			next.ServeHTTP(w, r)
		})
	}
}

func parsePublicKey(keyStr string) (ed25519.PublicKey, error) {
	var keyBytes []byte
	var err error

	// Try Base64 decode first
	keyBytes, err = base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		keyBytes = []byte(keyStr)
	}

	// 1. Try OpenSSH format
	if bytes.Contains(keyBytes, []byte("ssh-ed25519")) || bytes.Contains(keyBytes, []byte("BEGIN")) {
		pub, _, _, _, err := ssh.ParseAuthorizedKey(keyBytes)
		if err == nil {
			if edKey, ok := pub.(ssh.CryptoPublicKey); ok {
				if pk, ok := edKey.CryptoPublicKey().(ed25519.PublicKey); ok {
					return pk, nil
				}
			}
		}

		// Try as raw SSH body (sometimes folks copy just the base64 part of the pubkey)
		// But usually it's easier to try parsing as a generic PEM PKIX public key
		block, _ := pem.Decode(keyBytes)
		if block != nil {
			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				if pk, ok := pub.(ed25519.PublicKey); ok {
					return pk, nil
				}
			}
		}
	}

	// 2. Handle raw bytes
	if len(keyBytes) == ed25519.PublicKeySize {
		return ed25519.PublicKey(keyBytes), nil
	}

	return nil, fmt.Errorf("unsupported public key format or invalid size (%d bytes)", len(keyBytes))
}
