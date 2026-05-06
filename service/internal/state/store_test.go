package state

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
)

func TestNewLoadsSeededDatabase(t *testing.T) {
	t.Cleanup(config.Reset)

	store := openTestStore(t)
	t.Cleanup(func() {
		_ = store.Close()
	})

	snapshot := store.Snapshot()
	if got := len(snapshot.Config.Services); got != 3 {
		t.Fatalf("expected 3 services, got %d", got)
	}
	if got := len(snapshot.Config.EmailAccounts); got != 2 {
		t.Fatalf("expected 2 email accounts, got %d", got)
	}
	if got := snapshot.Config.Sms.FortySixElks.Username; got != "api_user_id" {
		t.Fatalf("unexpected sms username: %q", got)
	}
	if got := len(snapshot.Messages); got != 2 {
		t.Fatalf("expected 2 seeded messages, got %d", got)
	}
	if got := len(snapshot.Activities); got != 3 {
		t.Fatalf("expected 3 seeded activities, got %d", got)
	}
}

func TestUpdateConfigPersistsAcrossReopen(t *testing.T) {
	t.Cleanup(config.Reset)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mds.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if err := store.UpdateConfig(func(cfg *config.Config) error {
		cfg.Debug = true
		cfg.AdminBearerToken = "updated-token"
		cfg.Server.Host = "127.0.0.1"
		cfg.Server.Port = 9090
		return nil
	}); err != nil {
		t.Fatalf("UpdateConfig() failed: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	reopened, err := New(dbPath)
	if err != nil {
		t.Fatalf("reopen New() failed: %v", err)
	}
	t.Cleanup(func() {
		_ = reopened.Close()
	})

	cfg := reopened.Snapshot().Config
	if !cfg.Debug {
		t.Fatal("expected debug to persist")
	}
	if cfg.AdminBearerToken != "updated-token" {
		t.Fatalf("unexpected admin bearer token: %q", cfg.AdminBearerToken)
	}
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != 9090 {
		t.Fatalf("unexpected server settings: %+v", cfg.Server)
	}
}

func TestAddMessagePersistsRecipients(t *testing.T) {
	t.Cleanup(config.Reset)

	store := openTestStore(t)
	t.Cleanup(func() {
		_ = store.Close()
	})

	now := time.Now().UTC()
	message := api.AdminMessage{
		Id:           "msg-test-1",
		Channel:      api.AdminMessageChannelEmail,
		ContentMode:  api.Plain,
		CreatedAt:    now,
		Recipients:   []string{"recipient@example.com", "backup@example.com"},
		Sender:       "support@example.com",
		ServiceId:    "billing-api",
		Status:       api.Queued,
		Subject:      stringPtr("Test subject"),
		Body:         stringPtr("Test body"),
		TemplateName: nil,
	}

	if err := store.AddMessage(message, `{"id":"msg-test-1"}`, "rendered preview"); err != nil {
		t.Fatalf("AddMessage() failed: %v", err)
	}

	messages := store.ListMessages(10, nil, nil)
	found := false
	for _, msg := range messages {
		if msg.Id == message.Id {
			found = true
			if len(msg.Recipients) != 2 {
				t.Fatalf("expected 2 recipients, got %d", len(msg.Recipients))
			}
			break
		}
	}
	if !found {
		t.Fatal("queued message was not returned from store")
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mds.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	return store
}

func stringPtr(v string) *string {
	return &v
}
