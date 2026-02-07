# TKT-002 - Schema Validation on Ingest Endpoint

## Metadata

- ID: `TKT-002`
- Status: `done`
- Owner: `francis`
- Estimate: `M`
- Area: `ingest`

## Goal

- Validate incoming telemetry payloads against `packages/schemas/agent-event-v0.schema.json`.

## Scope

- Add `POST /v1/events`.
- Parse JSON payload and run schema validation.
- Return structured validation errors with HTTP `400`.

## Out of Scope

- Persisting valid events.
- Auth/RBAC.

## Deliverables

- Endpoint handler with schema validation path.
- Validation utility wired to v0 schema file.

## Test Plan

- Send valid payload and expect HTTP `202`.
- Send payload missing required fields and expect HTTP `400`.
- Send payload with invalid enum and expect HTTP `400`.

## Acceptance Criteria

- All invalid payloads are rejected before storage.
- Error body identifies failing field path.
- Valid payload acceptance latency is under 100ms locally.

## Dependencies

- `TKT-001`

## Notes

- Keep validator engine swappable if we move to generated code later.
- Implemented `POST /v1/events` in `services/ingest/internal/httpserver/server.go`.
- Added schema validator in `services/ingest/internal/validation/validator.go` loading from `packages/schemas/agent-event-v0.schema.json`.
- Added structured 400 responses:
  - `error: invalid_json`
  - `error: validation_failed` with per-field `path` and `message`.
- Added tests:
  - `services/ingest/internal/httpserver/server_test.go`
  - `services/ingest/internal/validation/validator_test.go`
- Verification:
  - `cd services/ingest && GOCACHE=/tmp/go-build go test ./...`
