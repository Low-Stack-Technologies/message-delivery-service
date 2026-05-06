-- name: ListEmailAccounts :many
SELECT id, address, display_name, smtp_host, smtp_port, smtp_username, smtp_password, is_default, status, last_tested_at, created_at, updated_at
FROM email_accounts
ORDER BY is_default DESC, created_at DESC, id ASC;

-- name: DeleteAllEmailAccounts :exec
DELETE FROM email_accounts;

-- name: InsertEmailAccount :exec
INSERT INTO email_accounts (id, address, display_name, smtp_host, smtp_port, smtp_username, smtp_password, is_default, status, last_tested_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

