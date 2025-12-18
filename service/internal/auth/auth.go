package auth

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
)

func NewMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := r.Header.Get("X-Client-Id")
			timestampStr := r.Header.Get("X-Timestamp")
			authHeader := r.Header.Get("Authorization")

			if clientID == "" || timestampStr == "" || authHeader == "" {
				http.Error(w, "Missing authentication headers", http.StatusUnauthorized)
				return
			}

			// 1. Verify Timestamp (Replay Protection)
			timestamp, err := time.Parse(time.RFC3339, timestampStr)
			if err != nil {
				http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
				return
			}
			if time.Since(timestamp) > 5*time.Minute || time.Since(timestamp) < -5*time.Minute {
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
				http.Error(w, "Unknown Client ID", http.StatusUnauthorized)
				return
			}

			pubKeyBytes, err := base64.StdEncoding.DecodeString(service.PublicKey)
			if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
				http.Error(w, "Service public key is misconfigured", http.StatusInternalServerError)
				return
			}
			pubKey := ed25519.PublicKey(pubKeyBytes)

			// 3. Construct Canonical Request
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for later handlers

			// For simplicity in this implementation: Canonical = Method + Path + Timestamp + Body
			// In a production app, we'd hash the body first.
			canonical := r.Method + "\n" + r.URL.Path + "\n" + timestampStr + "\n" + string(bodyBytes)

			// 4. Verify Signature
			if len(authHeader) < 10 || authHeader[:10] != "Signature " {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			signatureB64 := authHeader[10:]
			signature, err := base64.StdEncoding.DecodeString(signatureB64)
			if err != nil {
				http.Error(w, "Invalid signature encoding", http.StatusUnauthorized)
				return
			}

			if !ed25519.Verify(pubKey, []byte(canonical), signature) {
				http.Error(w, "Signature verification failed", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
