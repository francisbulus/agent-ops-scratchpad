# TKT-009 - Local Stack and Smoke Test

## Metadata

- ID: `TKT-009`
- Status: `backlog`
- Owner: `unassigned`
- Estimate: `M`
- Area: `ops`

## Goal

- Make MVP reproducible locally with one command and a smoke-test checklist.

## Scope

- Add local orchestration for app + DB (+ worker if separate process).
- Document environment variables and startup steps.
- Add smoke-test script/checklist for critical endpoints and UI.

## Out of Scope

- Production deployment manifests.
- Full CI pipeline hardening.

## Deliverables

- `docker-compose` or equivalent local run setup.
- `README` section for local bring-up.
- Smoke-test script or markdown runbook.

## Test Plan

- Start stack on clean machine environment.
- Run smoke test and verify all checks pass.
- Confirm teardown leaves no orphan processes.

## Acceptance Criteria

- New contributor can start MVP locally in under 15 minutes.
- Smoke-test verifies ingest, metrics, failures, and dashboard rendering.
- Steps are documented and deterministic.

## Dependencies

- `TKT-001` through `TKT-007`

## Notes

- Keep setup simple; optimize for reliability over sophistication.
