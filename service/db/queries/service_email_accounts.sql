-- name: ListServiceEmailAccounts :many
SELECT service_id, email_account_id, created_at
FROM service_email_accounts
ORDER BY service_id ASC, email_account_id ASC;

-- name: DeleteAllServiceEmailAccounts :exec
DELETE FROM service_email_accounts;

-- name: InsertServiceEmailAccount :exec
INSERT INTO service_email_accounts (service_id, email_account_id, created_at)
VALUES (?, ?, ?);
