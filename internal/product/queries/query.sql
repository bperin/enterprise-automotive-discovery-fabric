-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: GetProductByIdentifier :one
SELECT p.* FROM products p
JOIN product_identifiers pi ON p.id = pi.product_id
WHERE pi.normalized = $1 LIMIT 1;

-- name: ListProducts :many
SELECT * FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpsertProduct :one
INSERT INTO products (
    id, manufacturer_id, canonical_name, description, category, brand, oem, attributes, publication, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
)
ON CONFLICT (id) DO UPDATE SET
    manufacturer_id = EXCLUDED.manufacturer_id,
    canonical_name = EXCLUDED.canonical_name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    brand = EXCLUDED.brand,
    oem = EXCLUDED.oem,
    attributes = EXCLUDED.attributes,
    publication = EXCLUDED.publication,
    updated_at = NOW()
RETURNING *;
