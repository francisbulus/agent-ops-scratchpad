# TKT-001 - Ingest Service Scaffold

## Metadata

- ID: `TKT-001`
- Status: `ready`
- Owner: `unassigned`
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
