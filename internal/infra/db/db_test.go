package db_test

import (
	"context"
	"path/filepath"
	"testing"

	"enterprise-search/internal/infra/db"
)

func TestAutoMigrateSQLite(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_search.db")

	database, err := db.NewSQLite("file:" + dbPath)
	if err != nil {
		t.Fatalf("failed to open and automigrate SQLite DB: %v", err)
	}
	defer database.Close()

	// Verify schema creation by querying sqlite_master
	var count int
	err = database.QueryRowContext(context.Background(),
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='products'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}

	if count != 1 {
		t.Errorf("expected table 'products' to exist after auto-migration, count = %d", count)
	}

	// Test UUID helpers
	id := db.NewScopedUUID("prod")
	if len(id) < 36 {
		t.Errorf("expected valid scoped UUID, got %s", id)
	}
}
