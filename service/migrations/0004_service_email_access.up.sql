-- +migrate Up
ALTER TABLE services ADD COLUMN email_access_mode TEXT NOT NULL DEFAULT 'all' CHECK (email_access_mode IN ('all', 'restricted'));

CREATE TABLE service_email_accounts (
    service_id TEXT NOT NULL,
    email_account_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (service_id, email_account_id),
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    FOREIGN KEY (email_account_id) REFERENCES email_accounts(id) ON DELETE CASCADE
);

CREATE INDEX idx_service_email_accounts_service_id ON service_email_accounts(service_id);
CREATE INDEX idx_service_email_accounts_email_account_id ON service_email_accounts(email_account_id);
