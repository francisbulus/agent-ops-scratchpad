# Ingest Service

Minimal ingest service scaffold for MVP tickets `TKT-001` through `TKT-003`.

## Run

```bash
cd services/ingest
go run ./cmd/ingest
```

## Run with Autoreload

```bash
cd ../..
air -c .air.ingest.toml
```

This watches `services/ingest` Go files and `packages/schemas/*.json`, then restarts the service on changes.

Install `air` (once):

```bash
go install github.com/air-verse/air@latest
```

Environment variables:

- `PORT` (default: `8080`)
- `APP_ENV` (default: `dev`)
- `LOG_LEVEL` (default: `info`, one of `debug|info|warn|error`)
- `SHUTDOWN_TIMEOUT` (default: `10s`)
- `SCHEMA_PATH` (default resolves to `packages/schemas/agent-event-v0.schema.json`)
- `DATABASE_URL` (required, postgres DSN for event persistence)

## Database Migration

Run before starting the service:

```bash
psql "$DATABASE_URL" -f services/ingest/migrations/001_create_agent_events.sql
```

## Endpoints

```bash
curl -sS http://localhost:8080/healthz
curl -sS http://localhost:8080/readyz
curl -sS "http://localhost:8080/v1/metrics/overview?window_hours=24"
```

```bash
curl -sS -X POST http://localhost:8080/v1/events \
  -H 'Content-Type: application/json' \
  -d '{
    "event_version":"v0",
    "event_id":"123e4567-e89b-12d3-a456-426614174000",
    "event_type":"run.started",
    "occurred_at":"2026-02-07T21:00:00Z",
    "tenant":{"tenant_id":"t1","workspace_id":"w1","project_id":"p1"},
    "run":{"run_id":"r1","agent_id":"a1","workflow_id":"wf1","status":"started"},
    "trace":{"trace_id":"tr1","span_id":"sp1"}
  }'
```

`POST /v1/events` returns:

- `202` for valid payloads (with `persisted: true|false` for idempotency visibility)
- `400` with structured validation errors for invalid payloads
- `500` when persistence fails

`GET /v1/metrics/overview` returns:

- `200` with aggregate metrics (`total_runs`, `success_rate`, `total_cost_usd`, `avg_latency_ms`)
- defaults to last 24h when `window_hours` is omitted
- supports optional filters: `tenant_id`, `workspace_id`, `project_id`, `agent_id`, `workflow_id`

## Tests

```bash
cd services/ingest
GOCACHE=/tmp/go-build go test ./...
```
