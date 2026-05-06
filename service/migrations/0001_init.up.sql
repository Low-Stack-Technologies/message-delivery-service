-- +migrate Up
PRAGMA foreign_keys = ON;

CREATE TABLE app_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    server_host TEXT NOT NULL,
    server_port INTEGER NOT NULL,
    debug BOOLEAN NOT NULL DEFAULT 0,
    admin_bearer_token TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    owner TEXT NOT NULL DEFAULT '',
    scope TEXT NOT NULL CHECK (scope IN ('all', 'email', 'sms')),
    status TEXT NOT NULL CHECK (status IN ('active', 'paused')),
    public_key TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL,
    last_reroll_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE email_accounts (
    id TEXT PRIMARY KEY,
    address TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    smtp_host TEXT NOT NULL,
    smtp_port INTEGER NOT NULL,
    smtp_username TEXT NOT NULL,
    smtp_password TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('healthy', 'warning', 'offline')),
    last_tested_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_email_accounts_default ON email_accounts(is_default) WHERE is_default = 1;
CREATE UNIQUE INDEX idx_email_accounts_address ON email_accounts(address);

CREATE TABLE sms_credentials (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('connected', 'stale')),
    last_synced_at TIMESTAMP NOT NULL,
    rotation_count INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    channel TEXT NOT NULL CHECK (channel IN ('email', 'sms')),
    service_id TEXT NOT NULL,
    sender TEXT NOT NULL,
    subject TEXT,
    content_mode TEXT NOT NULL CHECK (content_mode IN ('plain', 'html', 'template')),
    body TEXT,
    template_name TEXT,
    request_json TEXT NOT NULL,
    rendered_text TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('queued', 'accepted', 'pending', 'sending', 'sent', 'failed')),
    created_at TIMESTAMP NOT NULL,
    queued_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    provider_message_id TEXT,
    provider_response_json TEXT,
    error_code TEXT,
    error_message TEXT,
    cost REAL,
    currency TEXT
);

CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_service_id ON messages(service_id, created_at DESC);
CREATE INDEX idx_messages_channel ON messages(channel, created_at DESC);
CREATE INDEX idx_messages_status ON messages(status, created_at DESC);

CREATE TABLE message_recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    ordinal INTEGER NOT NULL,
    recipient TEXT NOT NULL,
    recipient_name TEXT,
    country_code TEXT,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    UNIQUE (message_id, ordinal)
);

CREATE INDEX idx_message_recipients_message_id ON message_recipients(message_id, ordinal);

CREATE TABLE activity_log (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    detail TEXT NOT NULL,
    tone TEXT NOT NULL CHECK (tone IN ('info', 'success', 'warning', 'danger')),
    entity_type TEXT,
    entity_id TEXT,
    metadata_json TEXT,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_activity_log_created_at ON activity_log(created_at DESC);

