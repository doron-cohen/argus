package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/doron-cohen/argus/backend/api"
	"github.com/doron-cohen/argus/backend/internal/config"
	"github.com/doron-cohen/argus/backend/internal/health"
	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/go-chi/chi/v5"
)

func Start(cfg config.Config) (stop func(), err error) {
	mux := chi.NewRouter()

	// Mount healthz
	mux.Get("/healthz", health.HealthHandler)

	// Connect to PostgreSQL using storage.ConnectAndMigrate
	dsn := cfg.Storage.DSN()
	repo, dberr := storage.ConnectAndMigrate(context.Background(), dsn)
	if dberr != nil {
		slog.Error("Failed to connect or migrate database", "error", dberr)
		return nil, dberr
	}

	// Mount OpenAPI-generated handlers under /api
	mux.Mount("/api", api.Handler(api.NewAPIServer(repo)))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		slog.Info("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
		}
	}()

	stop = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Graceful shutdown failed", "error", err)
		}
	}

	return stop, nil
}
