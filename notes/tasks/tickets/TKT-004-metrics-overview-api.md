# TKT-004 - Metrics Overview API

## Metadata

- ID: `TKT-004`
- Status: `backlog`
- Owner: `unassigned`
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
