package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/francisbulus/agent-ops/services/ingest/internal/app"
	"github.com/francisbulus/agent-ops/services/ingest/internal/config"
	"github.com/francisbulus/agent-ops/services/ingest/internal/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logging.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	if err := app.Run(context.Background(), cfg, logger, sigCh); err != nil {
		logger.Error("service_exit_error", "error", err.Error())
		os.Exit(1)
	}
}
