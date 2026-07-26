-- +goose Up
-- SQL migration for Postgres: Add dealers, vehicles, marketing_content, and write authority tables

CREATE TABLE IF NOT EXISTS brands (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS vehicle_models (
    id VARCHAR(64) PRIMARY KEY,
    brand_id VARCHAR(64) REFERENCES brands(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(64) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS vehicle_trims (
    id VARCHAR(64) PRIMARY KEY,
    model_id VARCHAR(64) REFERENCES vehicle_models(id) ON DELETE CASCADE,
    model_year INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    msrp NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD' NOT NULL,
    attributes JSONB
);

CREATE TABLE IF NOT EXISTS dealers (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    service_enabled BOOLEAN DEFAULT true NOT NULL,
    parts_enabled BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS dealer_locations (
    id VARCHAR(64) PRIMARY KEY,
    dealer_id VARCHAR(64) REFERENCES dealers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address JSONB NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    postal_code VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS marketing_content (
    id VARCHAR(64) PRIMARY KEY,
    brand_id VARCHAR(64) NOT NULL,
    content_type VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(32) NOT NULL,
    effective_from TIMESTAMP WITH TIME ZONE,
    effective_to TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS marketing_content;
DROP TABLE IF EXISTS dealer_locations;
DROP TABLE IF EXISTS dealers;
DROP TABLE IF EXISTS vehicle_trims;
DROP TABLE IF EXISTS vehicle_models;
DROP TABLE IF EXISTS brands;
