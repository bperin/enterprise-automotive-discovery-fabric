-- +goose Up
-- SQL migration for SQLite: Add dealers, vehicles, marketing_content, and write authority tables

CREATE TABLE IF NOT EXISTS brands (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS vehicle_models (
    id TEXT PRIMARY KEY,
    brand_id TEXT REFERENCES brands(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS vehicle_trims (
    id TEXT PRIMARY KEY,
    model_id TEXT REFERENCES vehicle_models(id) ON DELETE CASCADE,
    model_year INTEGER NOT NULL,
    name TEXT NOT NULL,
    msrp REAL NOT NULL,
    currency TEXT DEFAULT 'USD' NOT NULL,
    attributes TEXT
);

CREATE TABLE IF NOT EXISTS dealers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    service_enabled INTEGER DEFAULT 1 NOT NULL,
    parts_enabled INTEGER DEFAULT 1 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS dealer_locations (
    id TEXT PRIMARY KEY,
    dealer_id TEXT REFERENCES dealers(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    postal_code TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS marketing_content (
    id TEXT PRIMARY KEY,
    brand_id TEXT NOT NULL,
    content_type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL,
    effective_from TIMESTAMP,
    effective_to TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS marketing_content;
DROP TABLE IF EXISTS dealer_locations;
DROP TABLE IF EXISTS dealers;
DROP TABLE IF EXISTS vehicle_trims;
DROP TABLE IF EXISTS vehicle_models;
DROP TABLE IF EXISTS brands;
