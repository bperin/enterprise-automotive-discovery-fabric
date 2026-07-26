// Package main is the composition root for the GCP Automotive Discovery Fabric API.
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/config"
	"enterprise-search/internal/content"
	"enterprise-search/internal/dealer"
	"enterprise-search/internal/discovery"
	"enterprise-search/internal/evidence"
	"enterprise-search/internal/fitment"
	"enterprise-search/internal/infra/postgres"
	"enterprise-search/internal/inventory"
	"enterprise-search/internal/product"
	"enterprise-search/internal/search"
	"enterprise-search/internal/vehicle"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 1. Load Process Configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Warn("Falling back to default environment config", slog.String("error", err.Error()))
		cfg = &ConfigFallback
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Connect to Production Cloud SQL PostgreSQL Pool & Run Auto-Migrations
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = postgres.DefaultDatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	pgPool, err := postgres.NewPool(ctx, dbURL, logger)
	cancel()
	if err != nil {
		logger.Error("Cloud SQL connection pool initialization failed; falling back to SQLite in-memory", slog.String("error", err.Error()))
	} else {
		defer pgPool.Close()

		// Run embedded Goose auto-migrations
		mCtx, mCancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := pgPool.AutoMigrate(mCtx); err != nil {
			logger.Error("Goose auto-migration failed on Cloud SQL", slog.String("error", err.Error()))
		}
		mCancel()
	}

	// 3. OIDC / OAuth2 Authenticator
	authenticator := auth.NewMockAuthenticator()

	// 4. Domain Repositories
	productRepo := product.NewMemoryRepository()
	fitmentRepo := fitment.NewMemoryRepository()
	inventoryRepo := inventory.NewMemoryRepository()
	evidenceRepo := evidence.NewMemoryRepository()
	dealerRepo := dealer.NewMemoryRepository()
	vehicleRepo := vehicle.NewMemoryRepository()
	contentRepo := content.NewMemoryRepository()
	searchEngine := search.NewMemorySearchEngine()

	// 5. Application Services
	productSvc := product.NewService(productRepo)
	fitmentSvc := fitment.NewService(fitmentRepo)
	inventorySvc := inventory.NewService(inventoryRepo)
	evidenceSvc := evidence.NewService(evidenceRepo)
	searchSvc := search.NewService(searchEngine)
	discoverySvc := discovery.NewService(searchSvc)

	// 6. HTTP Handlers
	productHandler := product.NewHandler(productSvc)
	fitmentHandler := fitment.NewHandler(fitmentSvc)
	inventoryHandler := inventory.NewHandler(inventorySvc)
	evidenceHandler := evidence.NewHandler(evidenceSvc)
	searchHandler := search.NewHandler(searchSvc)
	discoveryHandler := discovery.NewHandler(discoverySvc)
	dealerHandler := dealer.NewHandler(dealerRepo)
	vehicleHandler := vehicle.NewHandler(vehicleRepo)
	contentHandler := content.NewHandler(contentRepo)

	// 7. Route Registration
	mux := http.NewServeMux()

	productHandler.RegisterRoutes(mux, authenticator)
	fitmentHandler.RegisterRoutes(mux, authenticator)
	inventoryHandler.RegisterRoutes(mux, authenticator)
	evidenceHandler.RegisterRoutes(mux, authenticator)
	searchHandler.RegisterRoutes(mux, authenticator)
	discoveryHandler.RegisterRoutes(mux, authenticator)
	dealerHandler.RegisterRoutes(mux, authenticator)
	vehicleHandler.RegisterRoutes(mux, authenticator)
	contentHandler.RegisterRoutes(mux, authenticator)

	// Health and Readiness probes
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := searchSvc.Health(r.Context()); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("GCP Automotive Discovery Fabric API started",
			slog.String("port", port),
			slog.String("database_host", "localhost:5432"),
			slog.String("database_name", "automotive_discovery"),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}
	logger.Info("Server gracefully stopped")
}

var ConfigFallback = config.Config{
	Profile:     config.ProfileLocal,
	HTTP:        config.HTTP{Address: ":8080", RequestTimeout: 30 * time.Second, BodyLimit: 1048576, ShutdownTimeout: 10 * time.Second},
	DatabaseURL: postgres.DefaultDatabaseURL,
}
