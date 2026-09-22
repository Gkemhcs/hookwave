-- name: CreateDeliveries :copyfrom

INSERT INTO deliveries (id,message_id,endpoint_id,max_attempts,created_at,updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAllDeliveriesByMessageID :many

SELECT * FROM deliveries WHERE message_id = $1;

-- name: GetDeliveryByID :one

SELECT * FROM deliveries WHERE id = $1 ;

-- name: DeleteDeliveryByID :exec

DELETE FROM deliveries WHERE id = $1 ;

