package auth

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/state"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
	"golang.org/x/crypto/ssh"
)

type adminUserContextKey struct{}

func NewMiddleware(store *state.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			if r.URL.Path == "/v3/admin/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(r.URL.Path, "/v3/admin/") {
				user, ok := validateAdminBearer(w, r, store)
				if !ok {
					return
				}
				next.ServeHTTP(w, r.WithContext(WithAdminUser(r.Context(), user)))
				return
			}

			if !validateSignatureRequest(w, r) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func WithAdminUser(ctx context.Context, user api.AdminUser) context.Context {
	return context.WithValue(ctx, adminUserContextKey{}, user)
}

func AdminUserFromContext(ctx context.Context) (api.AdminUser, bool) {
	user, ok := ctx.Value(adminUserContextKey{}).(api.AdminUser)
	return user, ok
}

func validateAdminBearer(w http.ResponseWriter, r *http.Request, store *state.Store) (api.AdminUser, bool) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing bearer token", http.StatusUnauthorized)
		return api.AdminUser{}, false
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if store != nil {
		if user, err := store.ValidateAdminSession(r.Context(), token); err == nil {
			return user, true
		}
	}

	expected := os.Getenv("ADMIN_BEARER_TOKEN")
	if expected != "" && token == expected {
		if store != nil {
			if users, err := store.ListAdminUsers(r.Context()); err == nil && len(users) > 0 {
				return users[0], true
			}
		}
		return api.AdminUser{}, true
	}

	http.Error(w, "Invalid bearer token", http.StatusUnauthorized)
	return api.AdminUser{}, false
}

func validateSignatureRequest(w http.ResponseWriter, r *http.Request) bool {
	clientID := r.Header.Get("X-Client-Id")
	timestampStr := r.Header.Get("X-Timestamp")
	authHeader := r.Header.Get("Authorization")

	config.DebugLog("[DEBUG] Auth Attempt - ClientID: %s, Timestamp: %s, Auth: %s", clientID, timestampStr, authHeader)

	if clientID == "" || timestampStr == "" || authHeader == "" {
		config.DebugLog("[DEBUG] Auth Failed - Missing headers")
		http.Error(w, "Missing authentication headers", http.StatusUnauthorized)
		return false
	}

	// 1. Verify Timestamp (Replay Protection)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		config.DebugLog("[DEBUG] Auth Failed - Invalid timestamp format: %v", err)
		http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
		return false
	}
	if time.Since(timestamp) > 5*time.Minute || time.Since(timestamp) < -5*time.Minute {
		config.DebugLog("[DEBUG] Auth Failed - Timestamp expired: diff=%v", time.Since(timestamp))
		http.Error(w, "Request timestamp expired or in the future", http.StatusUnauthorized)
		return false
	}

	// 2. Find Service & Public Key
	cfg := config.Get()
	if cfg == nil {
		http.Error(w, "Service configuration is not available", http.StatusInternalServerError)
		return false
	}
	var service *config.ServiceConfig
	for i := range cfg.Services {
		if cfg.Services[i].ID == clientID {
			service = &cfg.Services[i]
			break
		}
	}

	if service == nil {
		config.DebugLog("[DEBUG] Auth Failed - Unknown Client ID: %s", clientID)
		http.Error(w, "Unknown Client ID", http.StatusUnauthorized)
		return false
	}

	pubKey, err := parsePublicKey(service.PublicKey)
	if err != nil {
		config.DebugLog("[DEBUG] Auth Failed - Public key parsing error: %v", err)
		http.Error(w, "Service public key is misconfigured", http.StatusInternalServerError)
		return false
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
		return false
	}
	signatureB64 := authHeader[10:]
	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		config.DebugLog("[DEBUG] Auth Failed - Signature decode error: %v", err)
		http.Error(w, "Invalid signature encoding", http.StatusUnauthorized)
		return false
	}

	if !ed25519.Verify(pubKey, []byte(canonical), signature) {
		config.DebugLog("[DEBUG] Auth Failed - Ed25519 verification failed")
		http.Error(w, "Signature verification failed", http.StatusUnauthorized)
		return false
	}

	config.DebugLog("[DEBUG] Auth Success - ClientID: %s", clientID)
	return true
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
