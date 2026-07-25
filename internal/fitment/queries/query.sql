-- name: GetFitmentAssertion :one
SELECT * FROM fitment_assertions
WHERE id = $1 LIMIT 1;

-- name: ListFitmentsByProduct :many
SELECT * FROM fitment_assertions
WHERE product_id = $1
ORDER BY observed_at DESC;

-- name: ListFitmentsByVehicle :many
SELECT * FROM fitment_assertions
WHERE vehicle_configuration_id = $1
ORDER BY observed_at DESC;
