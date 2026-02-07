package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/francisbulus/agent-ops/services/ingest/internal/config"
	"github.com/francisbulus/agent-ops/services/ingest/internal/httpserver"
	"github.com/francisbulus/agent-ops/services/ingest/internal/persistence/postgres"
	"github.com/francisbulus/agent-ops/services/ingest/internal/validation"
)

type server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// Run starts the HTTP server and handles graceful shutdown on context cancel or process signals.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger, signals <-chan os.Signal) error {
	if logger == nil {
		logger = slog.Default()
	}

	validator, err := validation.NewEventValidator(cfg.SchemaPath)
	if err != nil {
		return fmt.Errorf("initialize event validator: %w", err)
	}
	store, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize event store: %w", err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			logger.Error("event_store_close_failed", slog.String("error", closeErr.Error()))
		}
	}()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           httpserver.NewHandler(logger, validator, store),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger = logger.With(
		slog.String("addr", srv.Addr),
		slog.String("env", cfg.Env),
		slog.String("schema_path", cfg.SchemaPath),
		slog.Bool("db_enabled", cfg.DatabaseURL != ""),
	)
	return runServer(ctx, logger, cfg.ShutdownTimeout, signals, srv)
}

func runServer(ctx context.Context, logger *slog.Logger, shutdownTimeout time.Duration, signals <-chan os.Signal, srv server) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Info("server_starting")
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("server_shutdown_requested", slog.String("reason", "context_cancelled"))
	case sig := <-signals:
		logger.Info("server_shutdown_requested", slog.String("reason", "signal"), slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if err := <-errCh; err != nil {
		return err
	}

	logger.Info("server_stopped")
	return nil
}
