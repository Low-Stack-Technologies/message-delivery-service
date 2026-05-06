-- name: ListMessages :many
SELECT id, channel, service_id, sender, subject, content_mode, body, template_name, request_json, rendered_text, status, created_at, queued_at, accepted_at, provider_message_id, provider_response_json, error_code, error_message, cost, currency
FROM messages
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: InsertMessage :exec
INSERT INTO messages (id, channel, service_id, sender, subject, content_mode, body, template_name, request_json, rendered_text, status, created_at, queued_at, accepted_at, provider_message_id, provider_response_json, error_code, error_message, cost, currency)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListMessageRecipientsByMessageID :many
SELECT id, message_id, ordinal, recipient, recipient_name, country_code
FROM message_recipients
WHERE message_id = ?
ORDER BY ordinal ASC;

-- name: InsertMessageRecipient :exec
INSERT INTO message_recipients (message_id, ordinal, recipient, recipient_name, country_code)
VALUES (?, ?, ?, ?, ?);

