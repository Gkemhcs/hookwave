-- name: CreateEnvironment :one
INSERT INTO environments (id, name, slug, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5)
RETURNING *;


-- name: DeleteEnvironment :one
DELETE FROM environments WHERE id = $1 RETURNING id;


-- name: ListEnvironments :many 
SELECT * FROM environments;


-- name: GetEnvironmentById :one 
SELECT * FROM environments WHERE id=$1 ;