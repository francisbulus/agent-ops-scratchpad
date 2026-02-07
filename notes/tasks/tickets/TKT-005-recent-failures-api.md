# TKT-005 - Recent Failures API

## Metadata

- ID: `TKT-005`
- Status: `backlog`
- Owner: `unassigned`
- Estimate: `S`
- Area: `api`

## Goal

- Provide the dashboard with the most recent failed runs and key debug fields.

## Scope

- Add `GET /v1/failures/recent`.
- Return latest failures with `run_id`, `agent_id`, `workflow_id`, `error_type`, `occurred_at`.
- Support `limit` parameter with safe bounds.

## Out of Scope

- Full trace explorer.
- Error grouping analytics.

## Deliverables

- Endpoint + SQL query for recent failures.

## Test Plan

- Seed mixed success/failure events and call endpoint.
- Verify ordering is newest-first.
- Verify `limit` is capped and validated.

## Acceptance Criteria

- Endpoint returns only failures.
- Ordering and limit behavior are correct.
- Response shape is stable for dashboard consumption.

## Dependencies

- `TKT-003`

## Notes

- Include `trace_id` when available for quick drilldown linkage later.
