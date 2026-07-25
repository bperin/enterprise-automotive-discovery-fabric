-- +goose Up
-- Migration: Add Agent Runs and Product Reviews

CREATE TABLE IF NOT EXISTS agent_runs (
    id VARCHAR(64) PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL DEFAULT '',
    query TEXT NOT NULL,
    plan_intent VARCHAR(64) NOT NULL DEFAULT 'general',
    trace JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_reviews (
    id VARCHAR(64) PRIMARY KEY,
    product_id VARCHAR(64) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    author VARCHAR(128) NOT NULL DEFAULT 'Anonymous',
    rating INT NOT NULL DEFAULT 5,
    title VARCHAR(256) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    verified_buyer BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_product_reviews_product ON product_reviews(product_id);

-- +goose Down
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS agent_runs;
