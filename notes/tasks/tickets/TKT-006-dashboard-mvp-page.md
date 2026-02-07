# TKT-006 - Dashboard MVP Page

## Metadata

- ID: `TKT-006`
- Status: `backlog`
- Owner: `unassigned`
- Estimate: `M`
- Area: `dashboard`

## Goal

- Build a single dashboard page that answers the MVP operator questions.

## Scope

- Render KPI cards for runs, success rate, spend, and latency.
- Render recent failures table.
- Poll API every 30-60 seconds with visible last refresh time.

## Out of Scope

- Multi-page navigation.
- Role-based UI differences.

## Deliverables

- Page in `apps/dashboard` connected to `metrics/overview` and `failures/recent`.
- Loading and error states for both API calls.

## Test Plan

- Run dashboard against local API and validate card values.
- Simulate API error and verify graceful UI fallback.
- Verify refresh updates data without page reload.

## Acceptance Criteria

- Operator can identify current health and spend from one screen.
- No uncaught runtime errors on empty datasets.
- Core components render properly on desktop and mobile.

## Dependencies

- `TKT-004`, `TKT-005`

## Notes

- Keep UI minimal and readable; design polish can follow after MVP demo.
