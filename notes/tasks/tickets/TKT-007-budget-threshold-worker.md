# TKT-007 - Budget Threshold Worker and Slack Alert

## Metadata

- ID: `TKT-007`
- Status: `backlog`
- Owner: `unassigned`
- Estimate: `M`
- Area: `worker`

## Goal

- Detect daily spend threshold breaches and send one actionable Slack alert.

## Scope

- Add periodic worker (every 5 minutes).
- Compute spend vs configured threshold by scope.
- Send Slack webhook message on breach.
- Deduplicate repeated alerts for same threshold window.

## Out of Scope

- Multi-channel alert routing.
- Policy-based budget blocking.

## Deliverables

- Worker process/module with config.
- Slack notifier integration.

## Test Plan

- Seed spend below threshold and verify no alert.
- Seed spend above threshold and verify one alert.
- Run worker repeatedly and verify dedupe behavior.

## Acceptance Criteria

- Alert fires within 5 minutes after crossing threshold.
- Alert payload includes scope, threshold, and current spend.
- Duplicate alerts are suppressed within the same day window.

## Dependencies

- `TKT-003`

## Notes

- Use idempotency key per `scope + date + threshold`.
