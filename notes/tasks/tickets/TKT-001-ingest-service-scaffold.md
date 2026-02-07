# TKT-001 - Ingest Service Scaffold

## Metadata

- ID: `TKT-001`
- Status: `done`
- Owner: `francis`
- Estimate: `S`
- Area: `ingest`

## Goal

- Stand up a minimal Go ingest service skeleton that can run locally and expose health endpoints.

## Scope

- Create `services/ingest` Go module and app entrypoint.
- Add config loading for port and environment.
- Add `GET /healthz` and `GET /readyz` endpoints.

## Out of Scope

- Event validation logic.
- Database integration.

## Deliverables

- Bootable ingest service in `services/ingest`.
- Basic startup README note or inline run command.

## Test Plan

- Start service and curl `GET /healthz`.
- Start service and curl `GET /readyz`.

## Acceptance Criteria

- Service starts without panics on default config.
- Both health endpoints return HTTP `200`.
- Graceful shutdown works on `SIGINT`.

## Dependencies

- `none`

## Notes

- Keep code intentionally minimal to unblock follow-on tickets.
- Implemented files:
  - `services/ingest/cmd/ingest/main.go`
  - `services/ingest/internal/app/app.go`
  - `services/ingest/internal/config/config.go`
  - `services/ingest/internal/httpserver/server.go`
  - `services/ingest/internal/logging/logging.go`
  - `services/ingest/README.md`
  - tests under `services/ingest/internal/*/*_test.go`
- Verification:
  - `cd services/ingest && GOCACHE=/tmp/go-build go test ./...`
