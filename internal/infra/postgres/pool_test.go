package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"enterprise-search/internal/infra/postgres"
)

func TestPostgresPool(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping PostgreSQL live connection test; set DATABASE_URL to run")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, dbURL, nil)
	if err != nil {
		t.Fatalf("failed connecting to postgres pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping failed: %v", err)
	}

	if err := pool.AutoMigrate(ctx); err != nil {
		t.Fatalf("auto-migration failed: %v", err)
	}
}
