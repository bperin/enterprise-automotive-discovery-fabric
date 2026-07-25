package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"enterprise-search/internal/infra/postgres/sqlc"
	"enterprise-search/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

// Default Cloud SQL connection string for GCP Automotive Discovery Fabric.
const DefaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/automotive_discovery?sslmode=disable"

// PoolConfig controls connection pooling parameters for production Cloud SQL / PostgreSQL.
type PoolConfig struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// Pool wraps pgxpool.Pool with production connection management, sqlc integration, and Goose migrations.
type Pool struct {
	*pgxpool.Pool
	logger  *slog.Logger
	Queries *sqlc.Queries
}

// NewPool initializes a production-grade PostgreSQL connection pool using the provided connection string URL.
func NewPool(ctx context.Context, dsn string, logger *slog.Logger) (*Pool, error) {
	if strings.TrimSpace(dsn) == "" {
		dsn = DefaultDatabaseURL
	}

	return NewPoolWithConfig(ctx, PoolConfig{
		DSN:               dsn,
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   15 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}, logger)
}

// NewPoolWithConfig creates a pgxpool instance with explicit connection pooling settings.
func NewPoolWithConfig(ctx context.Context, cfg PoolConfig, logger *slog.Logger) (*Pool, error) {
	if strings.TrimSpace(cfg.DSN) == "" {
		cfg.DSN = DefaultDatabaseURL
	}

	config, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres connection string: %w", err)
	}

	// Apply connection pool settings
	if cfg.MaxConns > 0 {
		config.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		config.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		config.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		config.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		config.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create database connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres database at %s: %w", config.ConnConfig.Host, err)
	}

	if logger != nil {
		logger.Info("connected to PostgreSQL database pool",
			slog.String("host", config.ConnConfig.Host),
			slog.Int("port", int(config.ConnConfig.Port)),
			slog.String("database", config.ConnConfig.Database),
			slog.Int("max_conns", int(config.MaxConns)),
			slog.Int("min_conns", int(config.MinConns)),
		)
	}

	p := &Pool{
		Pool:    pool,
		logger:  logger,
		Queries: sqlc.New(pool),
	}

	return p, nil
}

// AutoMigrate applies embedded Goose SQL migrations to the connected PostgreSQL database.
func (p *Pool) AutoMigrate(ctx context.Context) error {
	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect postgres: %w", err)
	}

	// Create stdlib *sql.DB bridge for goose runner using pgx driver
	db, err := goose.OpenDBWithDriver("pgx", p.Config().ConnString())
	if err != nil {
		return fmt.Errorf("open stdlib sql db for goose migration: %w", err)
	}
	defer db.Close()

	if err := goose.UpContext(ctx, db, "postgres"); err != nil {
		return fmt.Errorf("run goose up migrations: %w", err)
	}

	if p.logger != nil {
		p.logger.Info("auto-migrations completed successfully")
	}
	return nil
}
