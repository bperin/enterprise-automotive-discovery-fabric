-- +goose Up
-- PostgreSQL Schema for GCP Automotive Discovery Fabric

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(64) PRIMARY KEY,
    manufacturer_id VARCHAR(128) NOT NULL DEFAULT '',
    canonical_name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(64) NOT NULL DEFAULT 'generic',
    brand VARCHAR(128) NOT NULL DEFAULT '',
    oem BOOLEAN NOT NULL DEFAULT FALSE,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    publication VARCHAR(32) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);

CREATE TABLE IF NOT EXISTS product_identifiers (
    id BIGSERIAL PRIMARY KEY,
    product_id VARCHAR(64) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    type VARCHAR(64) NOT NULL,
    value VARCHAR(256) NOT NULL,
    normalized VARCHAR(256) NOT NULL,
    source_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_product_identifiers_norm ON product_identifiers(normalized);

CREATE TABLE IF NOT EXISTS vehicle_configurations (
    id VARCHAR(128) PRIMARY KEY,
    year INT NOT NULL,
    make VARCHAR(128) NOT NULL,
    model VARCHAR(128) NOT NULL,
    trim VARCHAR(128) NOT NULL DEFAULT '',
    engine VARCHAR(128) NOT NULL DEFAULT '',
    drivetrain VARCHAR(64) NOT NULL DEFAULT '',
    body_style VARCHAR(64) NOT NULL DEFAULT '',
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS fitment_assertions (
    id VARCHAR(64) PRIMARY KEY,
    product_id VARCHAR(64) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    vehicle_configuration_id VARCHAR(128) NOT NULL REFERENCES vehicle_configurations(id) ON DELETE CASCADE,
    compatibility VARCHAR(32) NOT NULL DEFAULT 'unknown',
    conditions JSONB NOT NULL DEFAULT '[]'::jsonb,
    source_id VARCHAR(64) NOT NULL DEFAULT '',
    source_type VARCHAR(64) NOT NULL DEFAULT '',
    source_record_id VARCHAR(128) NOT NULL DEFAULT '',
    authority VARCHAR(32) NOT NULL DEFAULT 'derived',
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    evidence_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verification_status VARCHAR(64) NOT NULL DEFAULT 'model_extracted'
);

CREATE INDEX IF NOT EXISTS idx_fitment_product ON fitment_assertions(product_id);

CREATE TABLE IF NOT EXISTS locations (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    postal_code VARCHAR(32) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS inventory_observations (
    id VARCHAR(64) PRIMARY KEY,
    product_id VARCHAR(64) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    location_id VARCHAR(64) NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    quantity INT,
    availability VARCHAR(32) NOT NULL DEFAULT 'unknown',
    price_amount DOUBLE PRECISION,
    price_currency VARCHAR(8) NOT NULL DEFAULT 'USD',
    observed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    received_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    authority VARCHAR(32) NOT NULL DEFAULT 'authoritative',
    confidence DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    verification_status VARCHAR(64) NOT NULL DEFAULT 'verified_by_source_contract',
    source_id VARCHAR(64) NOT NULL DEFAULT '',
    source_type VARCHAR(64) NOT NULL DEFAULT '',
    source_record_id VARCHAR(128) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_inventory_product_loc ON inventory_observations(product_id, location_id);

CREATE TABLE IF NOT EXISTS assets (
    id VARCHAR(64) PRIMARY KEY,
    uri TEXT NOT NULL,
    media_type VARCHAR(64) NOT NULL,
    sha256 VARCHAR(64) NOT NULL DEFAULT '',
    page_count INT,
    width INT,
    height INT,
    source_id VARCHAR(64) NOT NULL DEFAULT '',
    source_type VARCHAR(64) NOT NULL DEFAULT '',
    source_record_id VARCHAR(128) NOT NULL DEFAULT '',
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS evidence (
    id VARCHAR(64) PRIMARY KEY,
    asset_id VARCHAR(64) NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    page INT,
    region JSONB,
    text TEXT NOT NULL DEFAULT '',
    image_uri TEXT NOT NULL DEFAULT '',
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    extractor_name VARCHAR(128) NOT NULL DEFAULT '',
    extractor_version VARCHAR(32) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS claims (
    id VARCHAR(64) PRIMARY KEY,
    subject_type VARCHAR(64) NOT NULL,
    subject_id VARCHAR(128) NOT NULL,
    field_path VARCHAR(256) NOT NULL,
    value JSONB NOT NULL DEFAULT '{}'::jsonb,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    authority VARCHAR(32) NOT NULL DEFAULT 'derived',
    source_id VARCHAR(64) NOT NULL DEFAULT '',
    source_type VARCHAR(64) NOT NULL DEFAULT '',
    source_record_id VARCHAR(128) NOT NULL DEFAULT '',
    evidence_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    extraction_run_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attestations (
    id VARCHAR(64) PRIMARY KEY,
    claim_id VARCHAR(64) NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'quarantined',
    reason_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    notes TEXT NOT NULL DEFAULT '',
    evidence_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    validator_type VARCHAR(64) NOT NULL DEFAULT '',
    validator_id VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
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
