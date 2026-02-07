# TKT-003 - Postgres Event Persistence

## Metadata

- ID: `TKT-003`
- Status: `backlog`
- Owner: `unassigned`
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
