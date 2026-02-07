package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("SCHEMA_PATH", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("cfg.Port = %d, want 8080", cfg.Port)
	}
	if cfg.Env != "dev" {
		t.Fatalf("cfg.Env = %q, want dev", cfg.Env)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("cfg.LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("cfg.ShutdownTimeout = %v, want 10s", cfg.ShutdownTimeout)
	}
	if cfg.SchemaPath != "packages/schemas/agent-event-v0.schema.json" {
		t.Fatalf("cfg.SchemaPath = %q, want default schema path", cfg.SchemaPath)
	}
}

func TestLoadAppliesSchemaPathOverride(t *testing.T) {
	t.Setenv("SCHEMA_PATH", "/tmp/schema.json")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SchemaPath != "/tmp/schema.json" {
		t.Fatalf("cfg.SchemaPath = %q, want /tmp/schema.json", cfg.SchemaPath)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT")
	}
}

func TestLoadRejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "bad-timeout")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid SHUTDOWN_TIMEOUT")
	}
}
