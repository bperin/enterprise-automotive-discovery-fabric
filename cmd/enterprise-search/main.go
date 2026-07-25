package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/config"
	"enterprise-search/internal/discovery"
	"enterprise-search/internal/evidence"
	"enterprise-search/internal/fitment"
	"enterprise-search/internal/infra/postgres"
	"enterprise-search/internal/inventory"
	"enterprise-search/internal/product"
	"enterprise-search/internal/search"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			Profile:     config.ProfileLocal,
			HTTP:        config.HTTP{Address: ":8080", RequestTimeout: 30 * time.Second, BodyLimit: 1048576, ShutdownTimeout: 10 * time.Second},
			DatabaseURL: postgres.DefaultDatabaseURL,
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = postgres.DefaultDatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	pgPool, err := postgres.NewPool(ctx, dbURL, logger)
	cancel()
	if err != nil {
		logger.Error("PostgreSQL pool creation failed", slog.String("error", err.Error()))
	} else {
		defer pgPool.Close()

		mCtx, mCancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := pgPool.AutoMigrate(mCtx); err != nil {
			logger.Error("Goose auto-migration failed", slog.String("error", err.Error()))
		}
		mCancel()
	}

	authenticator := auth.NewMockAuthenticator()

	productRepo := product.NewMemoryRepository()
	fitmentRepo := fitment.NewMemoryRepository()
	inventoryRepo := inventory.NewMemoryRepository()
	evidenceRepo := evidence.NewMemoryRepository()
	searchEngine := search.NewMemorySearchEngine()

	productSvc := product.NewService(productRepo)
	fitmentSvc := fitment.NewService(fitmentRepo)
	inventorySvc := inventory.NewService(inventoryRepo)
	evidenceSvc := evidence.NewService(evidenceRepo)
	searchSvc := search.NewService(searchEngine)
	discoverySvc := discovery.NewService(searchSvc)

	productHandler := product.NewHandler(productSvc)
	fitmentHandler := fitment.NewHandler(fitmentSvc)
	inventoryHandler := inventory.NewHandler(inventorySvc)
	evidenceHandler := evidence.NewHandler(evidenceSvc)
	searchHandler := search.NewHandler(searchSvc)
	discoveryHandler := discovery.NewHandler(discoverySvc)

	mux := http.NewServeMux()

	productHandler.RegisterRoutes(mux, authenticator)
	fitmentHandler.RegisterRoutes(mux, authenticator)
	inventoryHandler.RegisterRoutes(mux, authenticator)
	evidenceHandler.RegisterRoutes(mux, authenticator)
	searchHandler.RegisterRoutes(mux, authenticator)
	discoveryHandler.RegisterRoutes(mux, authenticator)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
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

	go func() {
		fmt.Printf("enterprise-search listening on :%s\n", port)
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
	fmt.Println("Server gracefully stopped")
}
