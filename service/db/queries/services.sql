-- name: ListServices :many
SELECT id, name, owner, scope, email_access_mode, status, public_key, notes, created_at, last_reroll_at, updated_at
FROM services
ORDER BY created_at DESC, id ASC;

-- name: DeleteAllServices :exec
DELETE FROM services;

-- name: InsertService :exec
INSERT INTO services (id, name, owner, scope, email_access_mode, status, public_key, notes, created_at, last_reroll_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
