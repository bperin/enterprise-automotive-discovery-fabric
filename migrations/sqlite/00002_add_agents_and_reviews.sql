-- +goose Up
-- Migration: Add Agent Runs and Product Reviews for SQLite

CREATE TABLE IF NOT EXISTS agent_runs (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL DEFAULT '',
    query TEXT NOT NULL,
    plan_intent TEXT NOT NULL DEFAULT 'general',
    trace TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_reviews (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    author TEXT NOT NULL DEFAULT 'Anonymous',
    rating INT NOT NULL DEFAULT 5,
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    verified_buyer BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_product_reviews_product ON product_reviews(product_id);

-- +goose Down
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS agent_runs;
