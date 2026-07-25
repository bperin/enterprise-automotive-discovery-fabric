-- +goose Up
-- SQLite Schema for GCP Automotive Discovery Fabric

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    manufacturer_id TEXT NOT NULL DEFAULT '',
    canonical_name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'generic',
    brand TEXT NOT NULL DEFAULT '',
    oem BOOLEAN NOT NULL DEFAULT FALSE,
    attributes TEXT NOT NULL DEFAULT '{}',
    publication TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);

CREATE TABLE IF NOT EXISTS product_identifiers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    normalized TEXT NOT NULL,
    source_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_product_identifiers_norm ON product_identifiers(normalized);

CREATE TABLE IF NOT EXISTS vehicle_configurations (
    id TEXT PRIMARY KEY,
    year INT NOT NULL,
    make TEXT NOT NULL,
    model TEXT NOT NULL,
    trim TEXT NOT NULL DEFAULT '',
    engine TEXT NOT NULL DEFAULT '',
    drivetrain TEXT NOT NULL DEFAULT '',
    body_style TEXT NOT NULL DEFAULT '',
    attributes TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS fitment_assertions (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    vehicle_configuration_id TEXT NOT NULL REFERENCES vehicle_configurations(id) ON DELETE CASCADE,
    compatibility TEXT NOT NULL DEFAULT 'unknown',
    conditions TEXT NOT NULL DEFAULT '[]',
    source_id TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_record_id TEXT NOT NULL DEFAULT '',
    authority TEXT NOT NULL DEFAULT 'derived',
    confidence REAL NOT NULL DEFAULT 0.0,
    evidence_ids TEXT NOT NULL DEFAULT '[]',
    observed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verification_status TEXT NOT NULL DEFAULT 'model_extracted'
);

CREATE INDEX IF NOT EXISTS idx_fitment_product ON fitment_assertions(product_id);

CREATE TABLE IF NOT EXISTS locations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    latitude REAL NOT NULL DEFAULT 0.0,
    longitude REAL NOT NULL DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS inventory_observations (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    location_id TEXT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    quantity INT,
    availability TEXT NOT NULL DEFAULT 'unknown',
    price_amount REAL,
    price_currency TEXT NOT NULL DEFAULT 'USD',
    observed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    authority TEXT NOT NULL DEFAULT 'authoritative',
    confidence REAL NOT NULL DEFAULT 1.0,
    verification_status TEXT NOT NULL DEFAULT 'verified_by_source_contract',
    source_id TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_record_id TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_inventory_product_loc ON inventory_observations(product_id, location_id);

CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    uri TEXT NOT NULL,
    media_type TEXT NOT NULL,
    sha256 TEXT NOT NULL DEFAULT '',
    page_count INT,
    width INT,
    height INT,
    source_id TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_record_id TEXT NOT NULL DEFAULT '',
    ingested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS evidence (
    id TEXT PRIMARY KEY,
    asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    page INT,
    region TEXT,
    text TEXT NOT NULL DEFAULT '',
    image_uri TEXT NOT NULL DEFAULT '',
    content_hash TEXT NOT NULL DEFAULT '',
    extractor_name TEXT NOT NULL DEFAULT '',
    extractor_version TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS claims (
    id TEXT PRIMARY KEY,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    field_path TEXT NOT NULL,
    value TEXT NOT NULL DEFAULT '{}',
    confidence REAL NOT NULL DEFAULT 0.0,
    authority TEXT NOT NULL DEFAULT 'derived',
    source_id TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_record_id TEXT NOT NULL DEFAULT '',
    evidence_ids TEXT NOT NULL DEFAULT '[]',
    extraction_run_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attestations (
    id TEXT PRIMARY KEY,
    claim_id TEXT NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'quarantined',
    reason_codes TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    evidence_ids TEXT NOT NULL DEFAULT '[]',
    validator_type TEXT NOT NULL DEFAULT '',
    validator_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS attestations;
DROP TABLE IF EXISTS claims;
DROP TABLE IF EXISTS evidence;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS inventory_observations;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS fitment_assertions;
DROP TABLE IF EXISTS vehicle_configurations;
DROP TABLE IF EXISTS product_identifiers;
DROP TABLE IF EXISTS products;
