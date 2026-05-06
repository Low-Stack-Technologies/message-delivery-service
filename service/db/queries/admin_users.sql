-- name: ListAdminUsers :many
SELECT id, username, created_at, updated_at, last_login_at
FROM admin_users
ORDER BY created_at ASC, username ASC;

-- name: GetAdminUserPublicByID :one
SELECT id, username, created_at, updated_at, last_login_at
FROM admin_users
WHERE id = ?;

-- name: GetAdminUserByUsername :one
SELECT id, username, password_hash, totp_secret, created_at, updated_at, last_login_at
FROM admin_users
WHERE username = ?;

-- name: CountAdminUsers :one
SELECT COUNT(*) AS count
FROM admin_users;

-- name: CreateAdminUser :one
INSERT INTO admin_users (id, username, password_hash, totp_secret, created_at, updated_at, last_login_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id, username, created_at, updated_at, last_login_at;

-- name: UpdateAdminUserLastLogin :exec
UPDATE admin_users
SET last_login_at = ?
WHERE id = ?;

-- name: GetAdminSessionByTokenHash :one
SELECT token_hash, user_id, created_at, expires_at, last_used_at, revoked_at, user_agent, remote_addr
FROM admin_sessions
WHERE token_hash = ?;

-- name: CreateAdminSession :one
INSERT INTO admin_sessions (token_hash, user_id, created_at, expires_at, last_used_at, revoked_at, user_agent, remote_addr)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING token_hash, user_id, created_at, expires_at, last_used_at, revoked_at, user_agent, remote_addr;

-- name: TouchAdminSession :exec
UPDATE admin_sessions
SET last_used_at = ?
WHERE token_hash = ?;
