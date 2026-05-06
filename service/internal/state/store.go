package state

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	sqlcdb "github.com/Low-Stack-Technologies/message-delivery-service/internal/db/sqlc"
	"github.com/Low-Stack-Technologies/message-delivery-service/migrations"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
	_ "github.com/glebarez/sqlite"
	migrate "github.com/rubenv/sql-migrate"
)

type Snapshot struct {
	Config     *config.Config
	Activities []api.AdminActivity
	Messages   []api.AdminMessage
}

type Store struct {
	mu sync.RWMutex

	path string
	db   *sql.DB
	q    *sqlcdb.Queries
}

func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	migrationSource := &migrate.EmbedFileSystemMigrationSource{FileSystem: migrations.FS, Root: "."}
	if _, err := migrate.Exec(db, "sqlite3", migrationSource, migrate.Up); err != nil {
		_ = db.Close()
		return nil, err
	}

	store := &Store{
		path: path,
		db:   db,
		q:    sqlcdb.New(db),
	}

	cfg, err := store.loadConfig(context.Background())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.ensureDefaultAdminUser(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	config.Set(cfg)

	return store, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := s.loadConfig(context.Background())
	if err != nil {
		cfg = config.Get()
		if cfg == nil {
			cfg = &config.Config{}
		}
	}
	activities, _ := s.loadActivities(context.Background(), 50)
	messages, _ := s.loadMessages(context.Background(), 50, nil, nil)

	return Snapshot{
		Config:     cfg,
		Activities: activities,
		Messages:   messages,
	}
}

func (s *Store) UpdateConfig(mutator func(*config.Config) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadConfig(context.Background())
	if err != nil {
		return err
	}

	clone := cloneConfig(cfg)
	if err := mutator(clone); err != nil {
		return err
	}

	if err := s.persistConfig(context.Background(), clone); err != nil {
		return err
	}

	config.Set(clone)
	return nil
}

func (s *Store) AddActivity(activity api.AdminActivity) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.addActivity(context.Background(), activity)
}

func (s *Store) AddMessage(message api.AdminMessage, requestJSON, renderedText string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.addMessage(context.Background(), message, requestJSON, renderedText)
}

func (s *Store) ListActivities(limit int) []api.AdminActivity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, err := s.loadActivities(context.Background(), limit)
	if err != nil {
		return nil
	}
	return items
}

func (s *Store) ListMessages(limit int, channel *api.AdminMessageChannel, serviceID *string) []api.AdminMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, err := s.loadMessages(context.Background(), limit, channel, serviceID)
	if err != nil {
		return nil
	}
	return items
}

func (s *Store) loadConfig(ctx context.Context) (*config.Config, error) {
	settings, err := s.q.GetAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	services, err := s.q.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	serviceEmailAccounts, err := s.q.ListServiceEmailAccounts(ctx)
	if err != nil {
		return nil, err
	}
	emailAccounts, err := s.q.ListEmailAccounts(ctx)
	if err != nil {
		return nil, err
	}
	sms, err := s.q.GetSmsCredentials(ctx)
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{}
	cfg.Server.Host = settings.ServerHost
	cfg.Server.Port = int(settings.ServerPort)
	cfg.Debug = settings.Debug
	cfg.AdminBearerToken = settings.AdminBearerToken
	cfg.Services = make([]config.ServiceConfig, 0, len(services))
	allowedEmailIDsByService := make(map[string][]string)
	for _, row := range serviceEmailAccounts {
		allowedEmailIDsByService[row.ServiceID] = append(allowedEmailIDsByService[row.ServiceID], row.EmailAccountID)
	}
	for _, service := range services {
		cfg.Services = append(cfg.Services, config.ServiceConfig{
			ID:                     service.ID,
			Name:                   service.Name,
			Owner:                  service.Owner,
			Scope:                  service.Scope,
			EmailAccessMode:        service.EmailAccessMode,
			AllowedEmailAccountIDs: append([]string(nil), allowedEmailIDsByService[service.ID]...),
			Status:                 service.Status,
			PublicKey:              service.PublicKey,
			Notes:                  service.Notes,
			CreatedAt:              service.CreatedAt,
			LastRerollAt:           service.LastRerollAt,
		})
	}
	cfg.EmailAccounts = make([]config.EmailAccountConfig, 0, len(emailAccounts))
	for _, account := range emailAccounts {
		var lastTestedAt *time.Time
		if account.LastTestedAt != nil {
			t := *account.LastTestedAt
			lastTestedAt = &t
		}
		cfg.EmailAccounts = append(cfg.EmailAccounts, config.EmailAccountConfig{
			ID:           account.ID,
			Address:      account.Address,
			DisplayName:  account.DisplayName,
			IsDefault:    account.IsDefault,
			Status:       account.Status,
			CreatedAt:    account.CreatedAt,
			UpdatedAt:    account.UpdatedAt,
			LastTestedAt: lastTestedAt,
			SMTP: struct {
				Host     string
				Port     int
				Username string
				Password string
			}{
				Host:     account.SmtpHost,
				Port:     int(account.SmtpPort),
				Username: account.SmtpUsername,
				Password: account.SmtpPassword,
			},
		})
	}
	cfg.Sms.FortySixElks = config.FortySixElksConfig{
		Username:      sms.Username,
		Password:      sms.Password,
		Status:        sms.Status,
		LastSyncedAt:  sms.LastSyncedAt,
		RotationCount: int(sms.RotationCount),
	}

	return cfg, nil
}

func (s *Store) persistConfig(ctx context.Context, cfg *config.Config) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.q.WithTx(tx)

	now := time.Now().UTC()
	if err := q.UpsertAppSettings(ctx, sqlcdb.UpsertAppSettingsParams{
		ServerHost:       cfg.Server.Host,
		ServerPort:       int64(cfg.Server.Port),
		Debug:            cfg.Debug,
		AdminBearerToken: cfg.AdminBearerToken,
		UpdatedAt:        now,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := q.DeleteAllServices(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, service := range cfg.Services {
		if err := q.InsertService(ctx, sqlcdb.InsertServiceParams{
			ID:              service.ID,
			Name:            service.Name,
			Owner:           service.Owner,
			Scope:           normalizeServiceScope(service.Scope),
			EmailAccessMode: normalizeEmailAccessMode(service.EmailAccessMode),
			Status:          normalizeServiceStatus(service.Status),
			PublicKey:       service.PublicKey,
			Notes:           service.Notes,
			CreatedAt:       service.CreatedAt,
			LastRerollAt:    service.LastRerollAt,
			UpdatedAt:       now,
		}); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := q.DeleteAllEmailAccounts(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	normalizeEmailDefaults(cfg.EmailAccounts)
	for _, account := range cfg.EmailAccounts {
		if err := q.InsertEmailAccount(ctx, sqlcdb.InsertEmailAccountParams{
			ID:           account.ID,
			Address:      account.Address,
			DisplayName:  account.DisplayName,
			SmtpHost:     account.SMTP.Host,
			SmtpPort:     int64(account.SMTP.Port),
			SmtpUsername: account.SMTP.Username,
			SmtpPassword: account.SMTP.Password,
			IsDefault:    account.IsDefault,
			Status:       normalizeEmailStatus(account.Status),
			LastTestedAt: account.LastTestedAt,
			CreatedAt:    account.CreatedAt,
			UpdatedAt:    now,
		}); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := q.DeleteAllServiceEmailAccounts(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	emailAccountIDs := make(map[string]struct{}, len(cfg.EmailAccounts))
	for _, account := range cfg.EmailAccounts {
		emailAccountIDs[account.ID] = struct{}{}
	}
	for _, service := range cfg.Services {
		for _, accountID := range service.AllowedEmailAccountIDs {
			if _, ok := emailAccountIDs[accountID]; !ok {
				_ = tx.Rollback()
				return fmt.Errorf("service %q references unknown email account %q", service.ID, accountID)
			}
			if err := q.InsertServiceEmailAccount(ctx, sqlcdb.InsertServiceEmailAccountParams{
				ServiceID:      service.ID,
				EmailAccountID: accountID,
				CreatedAt:      now,
			}); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	if err := q.UpsertSmsCredentials(ctx, sqlcdb.UpsertSmsCredentialsParams{
		Username:      cfg.Sms.FortySixElks.Username,
		Password:      cfg.Sms.FortySixElks.Password,
		Status:        normalizeSmsStatus(cfg.Sms.FortySixElks.Status),
		LastSyncedAt:  cfg.Sms.FortySixElks.LastSyncedAt,
		RotationCount: int64(cfg.Sms.FortySixElks.RotationCount),
		UpdatedAt:     now,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) addActivity(ctx context.Context, activity api.AdminActivity) error {
	return s.q.InsertActivity(ctx, sqlcdb.InsertActivityParams{
		ID:           activity.Id,
		Title:        activity.Title,
		Detail:       activity.Detail,
		Tone:         string(activity.Tone),
		EntityType:   nil,
		EntityID:     nil,
		MetadataJson: nil,
		CreatedAt:    activity.CreatedAt,
	})
}

func (s *Store) addMessage(ctx context.Context, message api.AdminMessage, requestJSON, renderedText string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.q.WithTx(tx)

	acceptedAt := (*time.Time)(nil)
	if message.Status == api.Accepted {
		t := message.CreatedAt
		acceptedAt = &t
	}

	if err := q.InsertMessage(ctx, sqlcdb.InsertMessageParams{
		ID:                   message.Id,
		Channel:              string(message.Channel),
		ServiceID:            message.ServiceId,
		Sender:               message.Sender,
		Subject:              message.Subject,
		ContentMode:          string(message.ContentMode),
		Body:                 message.Body,
		TemplateName:         message.TemplateName,
		RequestJson:          requestJSON,
		RenderedText:         renderedText,
		Status:               string(message.Status),
		CreatedAt:            message.CreatedAt,
		QueuedAt:             message.CreatedAt,
		AcceptedAt:           acceptedAt,
		ProviderMessageID:    nil,
		ProviderResponseJson: nil,
		ErrorCode:            nil,
		ErrorMessage:         nil,
		Cost:                 nil,
		Currency:             nil,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	for idx, recipient := range message.Recipients {
		if err := q.InsertMessageRecipient(ctx, sqlcdb.InsertMessageRecipientParams{
			MessageID:     message.Id,
			Ordinal:       int64(idx),
			Recipient:     recipient,
			RecipientName: nil,
			CountryCode:   nil,
		}); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) loadActivities(ctx context.Context, limit int) ([]api.AdminActivity, error) {
	if limit <= 0 {
		return []api.AdminActivity{}, nil
	}
	rows, err := s.q.ListActivities(ctx, int64(limit))
	if err != nil {
		return nil, err
	}
	activities := make([]api.AdminActivity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, api.AdminActivity{
			Id:        row.ID,
			Title:     row.Title,
			Detail:    row.Detail,
			Tone:      api.AdminActivityTone(row.Tone),
			CreatedAt: row.CreatedAt,
		})
	}
	return activities, nil
}

func (s *Store) loadMessages(ctx context.Context, limit int, channel *api.AdminMessageChannel, serviceID *string) ([]api.AdminMessage, error) {
	if limit <= 0 {
		return []api.AdminMessage{}, nil
	}
	rows, err := s.q.ListMessages(ctx, int64(^uint64(0)>>1))
	if err != nil {
		return nil, err
	}
	items := make([]api.AdminMessage, 0, len(rows))
	for _, row := range rows {
		if channel != nil && row.Channel != string(*channel) {
			continue
		}
		if serviceID != nil && row.ServiceID != *serviceID {
			continue
		}

		recipients, err := s.q.ListMessageRecipientsByMessageID(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		recipientValues := make([]string, 0, len(recipients))
		for _, recipient := range recipients {
			recipientValues = append(recipientValues, recipient.Recipient)
		}

		items = append(items, api.AdminMessage{
			Id:           row.ID,
			Channel:      api.AdminMessageChannel(row.Channel),
			ServiceId:    row.ServiceID,
			Recipients:   recipientValues,
			Sender:       row.Sender,
			Subject:      row.Subject,
			ContentMode:  api.AdminMessageContentMode(row.ContentMode),
			Body:         row.Body,
			TemplateName: row.TemplateName,
			CreatedAt:    row.CreatedAt,
			Status:       api.AdminMessageStatus(row.Status),
		})

		if len(items) >= limit {
			break
		}
	}

	return items, nil
}

func cloneConfig(cfg *config.Config) *config.Config {
	if cfg == nil {
		return &config.Config{}
	}

	out := *cfg
	out.Services = append([]config.ServiceConfig(nil), cfg.Services...)
	for i := range out.Services {
		out.Services[i].AllowedEmailAccountIDs = append([]string(nil), cfg.Services[i].AllowedEmailAccountIDs...)
	}
	out.EmailAccounts = append([]config.EmailAccountConfig(nil), cfg.EmailAccounts...)
	return &out
}

func normalizeServiceScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "email", "sms", "all":
		return strings.ToLower(strings.TrimSpace(scope))
	case "":
		return "all"
	default:
		return "all"
	}
}

func normalizeEmailAccessMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "all", "restricted":
		return strings.ToLower(strings.TrimSpace(mode))
	case "":
		return "all"
	default:
		return "all"
	}
}

func normalizeServiceStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "paused":
		return strings.ToLower(strings.TrimSpace(status))
	case "":
		return "active"
	default:
		return "active"
	}
}

func normalizeEmailStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "healthy", "warning", "offline":
		return strings.ToLower(strings.TrimSpace(status))
	case "":
		return "healthy"
	default:
		return "healthy"
	}
}

func normalizeSmsStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "connected", "stale":
		return strings.ToLower(strings.TrimSpace(status))
	case "":
		return "connected"
	default:
		return "connected"
	}
}

func normalizeEmailDefaults(accounts []config.EmailAccountConfig) {
	defaultSeen := false
	for i := range accounts {
		if accounts[i].IsDefault && !defaultSeen {
			defaultSeen = true
			continue
		}
		accounts[i].IsDefault = false
	}
	if !defaultSeen && len(accounts) > 0 {
		accounts[0].IsDefault = true
	}
}
