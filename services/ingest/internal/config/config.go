package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultPort            = 8080
	defaultEnv             = "dev"
	defaultLogLevel        = "info"
	defaultShutdownTimeout = 10 * time.Second
)

// Config holds runtime settings for the ingest service.
type Config struct {
	Port            int
	Env             string
	LogLevel        string
	ShutdownTimeout time.Duration
}

// Load reads config from environment with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		Port:            defaultPort,
		Env:             defaultEnv,
		LogLevel:        defaultLogLevel,
		ShutdownTimeout: defaultShutdownTimeout,
	}

	if raw := os.Getenv("PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port <= 0 {
			return Config{}, fmt.Errorf("invalid PORT: %q", raw)
		}
		cfg.Port = port
	}

	if raw := os.Getenv("APP_ENV"); raw != "" {
		cfg.Env = raw
	}

	if raw := os.Getenv("LOG_LEVEL"); raw != "" {
		cfg.LogLevel = raw
	}

	if raw := os.Getenv("SHUTDOWN_TIMEOUT"); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil || timeout <= 0 {
			return Config{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %q", raw)
		}
		cfg.ShutdownTimeout = timeout
	}

	return cfg, nil
}
