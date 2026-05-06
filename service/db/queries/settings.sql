-- name: GetAppSettings :one
SELECT id, server_host, server_port, debug, admin_bearer_token, updated_at
FROM app_settings
WHERE id = 1;

-- name: UpsertAppSettings :exec
INSERT INTO app_settings (id, server_host, server_port, debug, admin_bearer_token, updated_at)
VALUES (1, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  server_host = excluded.server_host,
  server_port = excluded.server_port,
  debug = excluded.debug,
  admin_bearer_token = excluded.admin_bearer_token,
  updated_at = excluded.updated_at;

-- name: GetSmsCredentials :one
SELECT id, username, password, status, last_synced_at, rotation_count, updated_at
FROM sms_credentials
WHERE id = 1;

-- name: UpsertSmsCredentials :exec
INSERT INTO sms_credentials (id, username, password, status, last_synced_at, rotation_count, updated_at)
VALUES (1, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  username = excluded.username,
  password = excluded.password,
  status = excluded.status,
  last_synced_at = excluded.last_synced_at,
  rotation_count = excluded.rotation_count,
  updated_at = excluded.updated_at;

