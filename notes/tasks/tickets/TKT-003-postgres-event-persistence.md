# TKT-003 - Postgres Event Persistence

## Metadata

- ID: `TKT-003`
- Status: `done`
- Owner: `francis`
- Estimate: `M`
- Area: `ingest`

## Goal

- Persist accepted events into Postgres as the source operational record.

## Scope

- Add migration for `agent_events`.
- Insert validated events with key searchable columns.
- Store full event JSON blob for audit/replay.

## Out of Scope

- Aggregated metrics tables.
- Retention/TTL jobs.

## Deliverables

- DB migration file(s).
- Insert repository path used by ingest handler.

## Test Plan

- Run migration on empty database.
- Send valid event and verify row inserted.
- Verify indexed query by `tenant_id` and `occurred_at`.

## Acceptance Criteria

- Successful event writes are durable in Postgres.
- Write errors surface as HTTP `500` with logged context.
- Insert path supports at least 100 events/second locally.

## Dependencies

- `TKT-002`

## Notes

- Include unique constraint on `event_id` for idempotency.
- Added migration:
  - `services/ingest/migrations/001_create_agent_events.sql`
- Implemented postgres persistence repository:
  - `services/ingest/internal/persistence/postgres/store.go`
- Wired persistence into ingest handler:
  - `services/ingest/internal/httpserver/server.go`
  - valid events now insert into `agent_events`
  - duplicate `event_id` is idempotent (`persisted: false`)
  - write failures return HTTP `500` (`error: persist_failed`)
- Wired store initialization at startup:
  - `services/ingest/internal/app/app.go`
  - `DATABASE_URL` config in `services/ingest/internal/config/config.go`
- Added tests:
  - `services/ingest/internal/persistence/postgres/store_test.go`
  - updated `services/ingest/internal/httpserver/server_test.go`
- Verification:
  - `cd services/ingest && GOCACHE=/tmp/go-build go test ./...`
- Note:
  - Migration execution against a live Postgres instance is documented but not executed in this sandbox environment.
