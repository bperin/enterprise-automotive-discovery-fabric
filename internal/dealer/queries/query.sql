-- name: GetDealerByID :one
SELECT * FROM dealers
WHERE id = $1 LIMIT 1;

-- name: ListDealers :many
SELECT * FROM dealers
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: FindLocationsByPostalCode :many
SELECT * FROM dealer_locations
WHERE postal_code = $1;
