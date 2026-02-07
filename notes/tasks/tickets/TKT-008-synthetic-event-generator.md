# TKT-008 - Synthetic Event Generator

## Metadata

- ID: `TKT-008`
- Status: `backlog`
- Owner: `unassigned`
- Estimate: `S`
- Area: `ops`

## Goal

- Generate realistic synthetic events to test ingest, metrics, and alert flows end-to-end.

## Scope

- Build CLI script/tool to emit event batches.
- Support configurable success/failure mix and token/cost profiles.
- Include scenario that intentionally breaches budget threshold.

## Out of Scope

- Full load-testing framework.
- Production traffic replay.

## Deliverables

- Script in repo (location decided by implementation).
- Example command presets in docs.

## Test Plan

- Send 1,000 events and verify ingest accepts majority with expected schema validity.
- Trigger failure-heavy mode and verify failures endpoint updates.
- Trigger cost spike mode and verify budget worker alert path.

## Acceptance Criteria

- Team can reproduce demo dataset in under 5 minutes.
- Generator outputs deterministic run summary counts.
- Script exits non-zero on transport or validation failures.

## Dependencies

- `TKT-002`, `TKT-003`

## Notes

- Use seed input for reproducible payload generation.
