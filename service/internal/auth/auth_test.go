package auth

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"golang.org/x/crypto/ssh"
)

func TestMiddleware_AdminBearerAndSignedRequests(t *testing.T) {
	t.Cleanup(config.Reset)
	t.Setenv("ADMIN_BEARER_TOKEN", "test-token")

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() failed: %v", err)
	}

	cfg := &config.Config{}
	cfg.Services = []config.ServiceConfig{
		{
			ID:        "billing-api",
			PublicKey: string(sshAuthorizedKey(pub)),
		},
	}
	config.Set(cfg)

	middleware := NewMiddleware(nil)

	t.Run("admin bearer", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequest(http.MethodGet, "/v3/admin/dashboard", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, got %d", rec.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/v3/admin/dashboard", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected success, got %d", rec.Code)
		}
	})

	t.Run("signature auth", func(t *testing.T) {
		body := []byte(`{"hello":"world"}`)
		timestamp := time.Now().UTC().Format(time.RFC3339)
		canonical := strings.Join([]string{
			http.MethodPost,
			"/v3/email",
			timestamp,
			fmt.Sprintf("%x", sha256.Sum256(body)),
		}, "\n")
		signature := ed25519.Sign(priv, []byte(canonical))

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequest(http.MethodPost, "/v3/email", bytes.NewReader(body))
		req.Header.Set("X-Client-Id", "billing-api")
		req.Header.Set("X-Timestamp", timestamp)
		req.Header.Set("Authorization", "Signature "+base64.StdEncoding.EncodeToString(signature))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected success, got %d", rec.Code)
		}
	})
}

func sshAuthorizedKey(pub ed25519.PublicKey) []byte {
	// The middleware accepts OpenSSH authorized-key format.
	authorized, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil
	}
	return ssh.MarshalAuthorizedKey(authorized)
}
