# TKT-004 - Metrics Overview API

## Metadata

- ID: `TKT-004`
- Status: `done`
- Owner: `francis`
- Estimate: `M`
- Area: `api`

## Goal

- Expose MVP dashboard aggregates for runs, success rate, spend, and latency.

## Scope

- Add `GET /v1/metrics/overview`.
- Query last 24h by default with optional scope filters.
- Return deterministic JSON response contract.

## Out of Scope

- Long-range trend analytics.
- Caching layer.

## Deliverables

- API handler + query implementation.
- Response schema snippet in docs/comments.

## Test Plan

- Seed DB with sample events and call endpoint.
- Verify run count, success rate, and spend values match fixture math.
- Verify empty dataset response returns zeros, not errors.

## Acceptance Criteria

- Endpoint returns within 300ms on local seeded dataset.
- Metrics are mathematically correct for provided fixtures.
- Response fields are stable and documented.

## Dependencies

- `TKT-003`

## Notes

- Keep SQL simple and explainable for first release.
- Implemented endpoint in current ingest runtime:
  - `GET /v1/metrics/overview` in `services/ingest/internal/httpserver/server.go`
- Added query layer:
  - `services/ingest/internal/persistence/postgres/metrics.go`
  - filters: `tenant_id`, `workspace_id`, `project_id`, `agent_id`, `workflow_id`
  - window default: 24h (`window_hours` query param optional)
- Added deterministic response contract type:
  - `services/ingest/internal/persistence/types.go`
- Added tests:
  - `services/ingest/internal/httpserver/server_test.go`
  - `services/ingest/internal/persistence/postgres/metrics_test.go`
- Verification:
  - `cd services/ingest && GOCACHE=/tmp/go-build go test ./...`
- Note:
  - Endpoint currently lives in ingest service for MVP speed; can be moved to `services/api` in a later refactor without changing response contract.
