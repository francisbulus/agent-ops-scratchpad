# AgentOps Architecture (WIP)

Status: `WIP`  
Last updated: `2026-02-07`

## 1. Purpose

Define the working architecture for an AgentOps platform that can:

- track usage (runs, steps, model calls, tokens),
- track cost and budget consumption in near real time,
- capture failures and reliability signals,
- enforce governance decisions and auditability,
- provide one operational dashboard for engineering, ops, and finance.

This document is intentionally iterative and should evolve as tickets ship.

## 2. Target Outcomes

- Operators can answer: what is running, what is failing, and why.
- Finance can answer: what did we spend, where, and against which budget.
- Product/engineering can answer: which workflows are expensive or unreliable.
- Security/governance can answer: what was blocked/allowed and by which policy.

## 3. Scope and Non-Goals

In scope for MVP:

- event ingestion with schema validation,
- event storage and basic querying,
- overview metrics and recent failures endpoints,
- budget threshold alerts,
- basic dashboard.

Out of scope for MVP:

- advanced anomaly ML,
- full chargeback/invoice automation,
- deep multi-region compliance controls,
- full distributed tracing backend rollout.

## 4. Architecture Principles

- Event-first: every meaningful action is emitted as telemetry.
- Traceability: all events are correlated through `trace_id` and run context.
- Financial correctness: cost records are reproducible from usage + price book.
- Append-only financial history: corrections are compensating records.
- Contract-first: services validate payloads against versioned schema.
- Progressive hardening: start simple, keep extension points explicit.

## 5. Core Primitives

These are the domain primitives the platform is built around.

1. Tenant
`tenant_id`, `workspace_id`, `project_id`

2. Run
`run_id`, `agent_id`, `workflow_id`, `status`, timing metadata

3. Step
`step_id`, `run_id`, `step_type`, `status`, `latency_ms`

4. Model Call
`model_call_id`, `run_id`, `step_id`, `provider`, `model`

5. Token Meter
`input_tokens`, `output_tokens`, `cached_tokens`, `total_tokens`

6. Price Book
versioned provider/model pricing metadata (`price_book_version`)

7. Cost Ledger
`ledger_entry_id`, `run_id`, `model_call_id`, `cost_usd`, `currency`, `reason`

8. Budget
`budget_id`, scope (`tenant|workspace|project|agent`), period, enforcement, result

9. Policy Decision
`policy_id`, `decision`, actor metadata, reason

10. Alert
`alert_id`, type, threshold, current_value, destination, `sent_at`

11. Trace
`trace_id`, `span_id`, `parent_span_id`

## 6. System Components

### 6.1 Ingest Service (`services/ingest`)

Responsibilities:

- receive telemetry (`POST /v1/events`),
- validate payloads against `packages/schemas/agent-event-v0.schema.json`,
- reject malformed or invalid events early,
- emit structured service logs.

Current state:

- implemented and running for `TKT-001` and `TKT-002`.

### 6.2 API Service (`services/api`)

Responsibilities:

- serve dashboard query endpoints,
- aggregate and return operational metrics,
- return recent failure slices and eventually trace drilldowns.

Current state:

- planned (`TKT-004`, `TKT-005`).

### 6.3 Dashboard (`apps/dashboard`)

Responsibilities:

- visualize KPI summary and failures list,
- support fast operational triage,
- provide refresh-aware, near-real-time view.

Current state:

- planned (`TKT-006`).

### 6.4 Budget Worker / Alert Engine

Responsibilities:

- compute spend-vs-threshold on schedule,
- emit `budget.threshold_hit` and alert notifications,
- dedupe repetitive alerts within window.

Current state:

- planned (`TKT-007`).

### 6.5 Shared Contracts (`packages`)

Responsibilities:

- schema versioning and compatibility discipline,
- shared source of truth for event contract.

Current state:

- `packages/schemas/agent-event-v0.schema.json` exists.

## 7. Data Architecture

MVP keeps three logical data zones:

1. Event Store (operational truth)
- raw validated events, queryable by tenant/time/run/trace.

2. Cost Ledger (financial truth)
- append-only cost records derived from token usage + pricing.

3. Metrics Aggregates (query speed)
- pre-aggregated counters for dashboard/API latency and cost snapshots.

For early MVP, these can live in one Postgres instance with separate tables/schemas.

## 8. Canonical Event Contract

Current canonical schema:

- `packages/schemas/agent-event-v0.schema.json`

Key envelope fields:

- `event_version`, `event_id`, `event_type`, `occurred_at`
- `tenant`, `run`, `trace`

Conditional sections by event type:

- `resource_usage` + `cost` for completed/failed run/model events,
- `budget` for `budget.threshold_hit`,
- `policy` for `policy.decision`,
- `alert` for `alert.emitted`,
- `error` for failed events.

## 9. Primary Flows

### 9.1 Event Ingestion Flow

1. producer sends event to `POST /v1/events`
2. ingest validates JSON + schema
3. invalid payload -> `400` with field-level errors
4. valid payload -> accept and persist (persistence is next ticket)
5. downstream metrics/cost paths consume stored events

### 9.2 Cost and Budget Flow

1. model call events include token usage + model/provider
2. price book version determines normalized USD cost
3. cost ledger entry created (append-only)
4. worker compares rolling spend to configured budget rules
5. threshold breach emits alert + notification

### 9.3 Governance Flow

1. policy engine evaluates sensitive action
2. policy decision event emitted with actor/rule/outcome
3. dashboard/audit view surfaces allow/block/escalate history

## 10. Integration Model

The platform should support integration via adapters.

### 10.1 LLM Provider Integrations

- capture provider request metadata, model identifier, token usage.
- normalize provider differences into canonical event schema.

### 10.2 Agent Runtime Integrations

- SDK/middleware hooks emit run/step/model/tool events.
- integration can be library-based or sidecar-based.

### 10.3 Notification Integrations

- initial: Slack webhook for alerts.
- next: generic webhooks, email, PagerDuty/Opsgenie.

### 10.4 Billing/Finance Integrations

- optional reconciliation imports from provider statements.
- compare external invoice totals against internal metering.

## 11. Operational Concerns

Reliability:

- idempotency on `event_id` at persistence layer,
- backpressure strategy for ingest spikes,
- retry policy with bounded attempts.

Security:

- API authn/authz for ingest and query endpoints,
- PII minimization/redaction policy for payload content,
- immutable audit trail for governance events.

Observability of the platform:

- service logs with trace correlation,
- ingest success/error/latency metrics,
- worker execution and alert-delivery metrics.

## 12. Deployment View (Initial)

Local MVP:

- ingest service + api service + postgres + dashboard.

Production target (near-term):

- stateless ingest/api replicas behind load balancer,
- managed Postgres,
- worker deployment for budget checks and alert dispatch,
- centralized logs and metrics backend.

## 13. Ticket Mapping

- `TKT-001`: ingest scaffold (done)
- `TKT-002`: schema validation endpoint (done)
- `TKT-003`: event persistence + idempotency
- `TKT-004`: overview metrics API
- `TKT-005`: recent failures API
- `TKT-006`: dashboard MVP
- `TKT-007`: budget threshold worker + Slack alerts
- `TKT-008`: synthetic event generator
- `TKT-009`: local stack and smoke test
- `TKT-010`: structured logging and trace propagation hardening

## 14. Open Decisions

- single binary vs separate ingest/api services for MVP runtime,
- exact Postgres schema partitioning strategy (raw vs aggregate tables),
- budget evaluation cadence and granularity,
- policy engine embedding vs external policy service,
- schema evolution policy (`v0` to `v1`) and backward compatibility.

## 15. Next Document Updates (Planned)

- add concrete Postgres table definitions for event store/cost ledger/aggregates,
- add API response contracts for metrics and failures endpoints,
- add sequence diagrams for ingest->persist->aggregate and budget alerting,
- add threat model and data retention policy sections.
