# Ingest Service

Minimal ingest service scaffold for MVP ticket `TKT-001`.

## Run

```bash
cd services/ingest
go run ./cmd/ingest
```

Environment variables:

- `PORT` (default: `8080`)
- `APP_ENV` (default: `dev`)
- `LOG_LEVEL` (default: `info`, one of `debug|info|warn|error`)
- `SHUTDOWN_TIMEOUT` (default: `10s`)

## Health Endpoints

```bash
curl -sS http://localhost:8080/healthz
curl -sS http://localhost:8080/readyz
```

Both endpoints return HTTP `200` and a small JSON status payload.
