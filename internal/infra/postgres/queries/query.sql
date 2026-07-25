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

-- name: InsertProductIdentifier :exec
INSERT INTO product_identifiers (
    product_id, type, value, normalized, source_id
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetVehicleConfig :one
SELECT * FROM vehicle_configurations
WHERE id = $1 LIMIT 1;

-- name: UpsertVehicleConfig :one
INSERT INTO vehicle_configurations (
    id, year, make, model, trim, engine, drivetrain, body_style, attributes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (id) DO UPDATE SET
    year = EXCLUDED.year,
    make = EXCLUDED.make,
    model = EXCLUDED.model,
    trim = EXCLUDED.trim,
    engine = EXCLUDED.engine,
    drivetrain = EXCLUDED.drivetrain,
    body_style = EXCLUDED.body_style,
    attributes = EXCLUDED.attributes
RETURNING *;

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

-- name: UpsertFitmentAssertion :one
INSERT INTO fitment_assertions (
    id, product_id, vehicle_configuration_id, compatibility, conditions, source_id, source_type, source_record_id, authority, confidence, evidence_ids, observed_at, verification_status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (id) DO UPDATE SET
    compatibility = EXCLUDED.compatibility,
    conditions = EXCLUDED.conditions,
    authority = EXCLUDED.authority,
    confidence = EXCLUDED.confidence,
    evidence_ids = EXCLUDED.evidence_ids,
    observed_at = EXCLUDED.observed_at,
    verification_status = EXCLUDED.verification_status
RETURNING *;

-- name: GetLocation :one
SELECT * FROM locations
WHERE id = $1 LIMIT 1;

-- name: UpsertLocation :one
INSERT INTO locations (
    id, name, postal_code, latitude, longitude
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    postal_code = EXCLUDED.postal_code,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude
RETURNING *;

-- name: GetInventoryObservation :one
SELECT * FROM inventory_observations
WHERE id = $1 LIMIT 1;

-- name: ListInventoryByProduct :many
SELECT * FROM inventory_observations
WHERE product_id = $1
ORDER BY observed_at DESC;

-- name: UpsertInventoryObservation :one
INSERT INTO inventory_observations (
    id, product_id, location_id, quantity, availability, price_amount, price_currency, observed_at, received_at, expires_at, authority, confidence, verification_status, source_id, source_type, source_record_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
ON CONFLICT (id) DO UPDATE SET
    quantity = EXCLUDED.quantity,
    availability = EXCLUDED.availability,
    price_amount = EXCLUDED.price_amount,
    price_currency = EXCLUDED.price_currency,
    observed_at = EXCLUDED.observed_at,
    received_at = EXCLUDED.received_at,
    expires_at = EXCLUDED.expires_at,
    authority = EXCLUDED.authority,
    confidence = EXCLUDED.confidence,
    verification_status = EXCLUDED.verification_status
RETURNING *;

-- name: GetAsset :one
SELECT * FROM assets
WHERE id = $1 LIMIT 1;

-- name: UpsertAsset :one
INSERT INTO assets (
    id, uri, media_type, sha256, page_count, width, height, source_id, source_type, source_record_id, ingested_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (id) DO UPDATE SET
    uri = EXCLUDED.uri,
    media_type = EXCLUDED.media_type,
    sha256 = EXCLUDED.sha256,
    page_count = EXCLUDED.page_count,
    width = EXCLUDED.width,
    height = EXCLUDED.height
RETURNING *;

-- name: GetEvidenceByID :one
SELECT * FROM evidence
WHERE id = $1 LIMIT 1;

-- name: InsertEvidence :one
INSERT INTO evidence (
    id, asset_id, page, region, text, image_uri, content_hash, extractor_name, extractor_version, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: InsertClaim :one
INSERT INTO claims (
    id, subject_type, subject_id, field_path, value, confidence, authority, source_id, source_type, source_record_id, evidence_ids, extraction_run_id, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: ListClaimsBySubject :many
SELECT * FROM claims
WHERE subject_id = $1
ORDER BY created_at DESC;

-- name: GetAttestationForClaim :one
SELECT * FROM attestations
WHERE claim_id = $1 LIMIT 1;

-- name: UpsertAttestation :one
INSERT INTO attestations (
    id, claim_id, status, reason_codes, notes, evidence_ids, validator_type, validator_id, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    reason_codes = EXCLUDED.reason_codes,
    notes = EXCLUDED.notes,
    evidence_ids = EXCLUDED.evidence_ids,
    validator_type = EXCLUDED.validator_type,
    validator_id = EXCLUDED.validator_id
RETURNING *;

-- name: InsertAgentRun :one
INSERT INTO agent_runs (
    id, session_id, query, plan_intent, trace, status, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetAgentRun :one
SELECT * FROM agent_runs
WHERE id = $1 LIMIT 1;

-- name: ListAgentRuns :many
SELECT * FROM agent_runs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: InsertProductReview :one
INSERT INTO product_reviews (
    id, product_id, author, rating, title, content, verified_buyer, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListProductReviews :many
SELECT * FROM product_reviews
WHERE product_id = $1
ORDER BY created_at DESC;
