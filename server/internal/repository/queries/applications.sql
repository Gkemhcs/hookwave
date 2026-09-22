-- name: CreateApplication :one

INSERT INTO applications(id,name,slug,environment_id,created_at,updated_at)
VALUES ($1,$2,$3,$4,$5,$6) 
RETURNING *;

-- name: GetAllApplicationsByEnvironmentID :many 

SELECT * FROM applications WHERE environment_id=$1;


-- name: GetApplicationByID :one 

SELECT * FROM applications WHERE id=$1  AND environment_id=$2;

-- name: DeleteApplicationByID :exec

DELETE  FROM applications WHERE id=$1 AND environment_id=$2;


-- name: UpdateApplicationById :one 

UPDATE applications
SET
    name=$1 ,
    slug=$2 ,
    updated_at=NOW()
WHERE id=$3 AND environment_id=$4
RETURNING *
;


