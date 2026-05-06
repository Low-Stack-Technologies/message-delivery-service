-- name: ListActivities :many
SELECT id, title, detail, tone, entity_type, entity_id, metadata_json, created_at
FROM activity_log
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: InsertActivity :exec
INSERT INTO activity_log (id, title, detail, tone, entity_type, entity_id, metadata_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

