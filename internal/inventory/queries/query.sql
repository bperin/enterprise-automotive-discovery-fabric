-- name: GetObservation :one
SELECT * FROM inventory_observations
WHERE id = $1 LIMIT 1;

-- name: FindObservationsByProduct :many
SELECT * FROM inventory_observations
WHERE product_id = $1
ORDER BY observed_at DESC;
