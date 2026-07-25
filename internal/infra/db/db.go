package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"enterprise-search/migrations"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// DB wraps an sql.DB connection with automatic Goose migration capabilities and UUID helpers.
type DB struct {
	*sql.DB
	driver string
	mu     sync.Mutex
}

// Open initializes a database connection (PostgreSQL or SQLite), tests the connection, and automigrates embedded Goose migrations.
func Open(driverName, dsn string) (*DB, error) {
	if driverName == "" {
		if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			driverName = "pgx"
		} else {
			driverName = "sqlite"
		}
	}

	if driverName == "sqlite" && dsn == "" {
		dsn = "file:enterprise_search.db?cache=shared&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	}

	rawDB, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database (%s): %w", driverName, err)
	}

	if err := rawDB.Ping(); err != nil {
		rawDB.Close()
		return nil, fmt.Errorf("failed to ping database (%s): %w", driverName, err)
	}

	gooseDriver := "postgres"
	if driverName == "sqlite" || driverName == "sqlite3" {
		gooseDriver = "sqlite3"
	}

	database := &DB{
		DB:     rawDB,
		driver: gooseDriver,
	}

	if err := database.AutoMigrate(context.Background()); err != nil {
		rawDB.Close()
		return nil, fmt.Errorf("auto-migration failed: %w", err)
	}

	return database, nil
}

// NewSQLite creates a new SQLite database instance, enables WAL/foreign keys, and automatically runs Goose migrations.
func NewSQLite(dsn string) (*DB, error) {
	return Open("sqlite", dsn)
}

// NewPostgres creates a new PostgreSQL database instance and automatically runs Goose migrations.
func NewPostgres(dsn string) (*DB, error) {
	return Open("pgx", dsn)
}

// NewUUID generates a fresh v4 UUID string.
func NewUUID() string {
	return uuid.New().String()
}

// NewScopedUUID generates a prefixed UUID string (e.g. "prod-550e8400-e29b-41d4-a716-446655440000").
func NewScopedUUID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String())
}

// AutoMigrate runs embedded Goose database migrations automatically on application startup.
func (d *DB) AutoMigrate(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect(d.driver); err != nil {
		return fmt.Errorf("failed to set goose dialect %s: %w", d.driver, err)
	}

	migrationDir := "postgres"
	if d.driver == "sqlite3" {
		migrationDir = "sqlite"
	}

	if err := goose.UpContext(ctx, d.DB, migrationDir); err != nil {
		return fmt.Errorf("failed to run goose migrations up in %s: %w", migrationDir, err)
	}

	return nil
}
