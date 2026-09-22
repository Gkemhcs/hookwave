-- name: CreateMessage :one

INSERT INTO messages (id, application_id, event_type, payload, metadata, idempotency_key,created_at)
VALUES ($1, $2, $3, $4, $5, $6,$7)
RETURNING *;

-- name: GetAllMessagesByApplicationID :many

SELECT * FROM messages WHERE application_id = $1;

-- name: GetMessageByID :one

SELECT * FROM messages WHERE id = $1 AND application_id = $2;

-- name: DeleteMessageByID :exec

DELETE FROM messages WHERE id = $1 AND application_id = $2;
