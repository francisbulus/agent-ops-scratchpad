# TKT-010 - Structured Logs and Trace Context Propagation

## Metadata

- ID: `TKT-010`
- Status: `backlog`
- Owner: `unassigned`
- Estimate: `S`
- Area: `api`

## Goal

- Ensure requests and event processing are diagnosable with structured logs and trace identifiers.

## Scope

- Standardize JSON logging fields across ingest/api paths.
- Propagate `trace_id` and `span_id` through request lifecycle.
- Include error classification and request latency in logs.

## Out of Scope

- Full distributed tracing backend integration.
- Log retention pipeline.

## Deliverables

- Logging middleware/utilities.
- Updated handlers emitting structured logs with trace fields.

## Test Plan

- Send request with trace headers and verify logs include propagated IDs.
- Trigger validation failure and verify error log contains type and field path.
- Verify latency field present for successful requests.

## Acceptance Criteria

- Logs are machine-parseable JSON.
- Every request log includes correlation identifiers.
- Failure logs are actionable without reproducing locally.

## Dependencies

- `TKT-001`, `TKT-002`

## Notes

- Align log key names with event schema (`trace_id`, `span_id`).
