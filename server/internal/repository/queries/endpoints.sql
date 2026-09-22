-- name: CreateEndpoint :one

INSERT INTO endpoints (id, name, application_id, url, method, headers, query_params, event_types, signing_secret, is_active, timeout_ms, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetAllEndpointsByApplicationID :many

SELECT * FROM endpoints WHERE application_id = $1;

-- name: GetEndpointByID :one

SELECT * FROM endpoints WHERE id = $1 AND application_id = $2;

-- name: DeleteEndpointByID :exec

DELETE FROM endpoints WHERE id = $1 AND application_id = $2;

-- name: UpdateEndpointByID :one

UPDATE endpoints
SET
    name = $1,
    url = $2,
    method = $3,
    headers = $4,
    query_params = $5,
    event_types = $6,
    signing_secret = $7,
    is_active = $8,
    timeout_ms = $9,
    updated_at = NOW()
WHERE id=$10 and application_id=$11
RETURNING *;

-- name: UpdateEndpointStatus :one

UPDATE endpoints
SET
    is_active = $1,
    updated_at = NOW()
WHERE id = $2 AND application_id = $3
RETURNING *;
