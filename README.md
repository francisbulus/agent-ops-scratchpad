# AgentOps Dashboard Scratch Pad

Working ideas for an agent observability + governance dashboard to track cost, usage, billing, failures, and reliability.

## Core Outcomes

- See total spend in real time and by tenant/team/model.
- Detect runaway cost or token spikes early.
- Track reliability (timeouts, tool failures, retries, degraded output quality).
- Support governance (audit trails, policy checks, approvals, access controls).
- Make incident response fast (what failed, why, where, and impact).

## Key Questions

- Which agents are worth their cost?
- Which workflows fail most often and at what step?
- Are failures provider/model/tool-specific or prompt/version-specific?
- Where are we leaking money (retries, loops, oversized contexts, unnecessary tools)?
- Which customers/teams are near budget limits?

## Metrics to Capture

### Cost & Billing

- Total spend: daily/weekly/monthly.
- Spend by: org, workspace, project, agent, workflow, model, provider.
- Unit economics: cost per successful task, cost per user action, cost per API call.
- Token metrics: prompt/completion/total; cache hit ratio if available.
- Budget variance: actual vs planned budget.
- Forecast: projected month-end spend.

### Usage

- Request volume: RPM/RPH and unique active users/agents.
- Session counts and average session length.
- Tool call volume per tool and per workflow step.
- Concurrency and queue depth.
- P50/P90/P99 latency by step/provider/model.

### Reliability & Failures

- Success rate (end-to-end and per step).
- Error rate by type: timeout, auth, quota, validation, tool runtime, policy block.
- Retry rate and retry success rate.
- Fallback activation rate (model fallback, tool fallback).
- Hallucination/quality guardrail violation rate (if scored).

### Governance & Security

- Policy decision logs: allow/block/escalate.
- Sensitive action audit log (who did what, when, with what context).
- PII detection/redaction counts.
- Data egress destinations by workflow.
- Human approval queue metrics and SLA.

## Event Model (Telemetry)

### Event Types

- `agent.request.started`
- `agent.request.completed`
- `agent.step.started`
- `agent.step.completed`
- `agent.tool.called`
- `agent.tool.failed`
- `agent.retry`
- `agent.fallback.triggered`
- `agent.policy.decision`
- `agent.budget.threshold_hit`

### Required Event Fields

- `timestamp`
- `trace_id`, `span_id`, `parent_span_id`
- `org_id`, `workspace_id`, `project_id`
- `user_id` (or service principal)
- `agent_id`, `workflow_id`, `workflow_version`
- `provider`, `model`
- `prompt_version`
- `status`, `error_type`, `error_message_hash`
- `latency_ms`
- `tokens_prompt`, `tokens_completion`, `tokens_total`
- `cost_usd`
- `tool_name` (if applicable)
- `policy_id`, `policy_result` (if applicable)

## Dashboard Views

### 1) Executive Overview

- Spend today vs budget.
- Success rate and incident count.
- Top 5 expensive workflows.
- Forecasted month-end overrun risk.

### 2) Operations / SRE View

- Live requests + failure heatmap by step.
- Latency histograms by model/tool/provider.
- Error taxonomy panel with drilldown.
- Retry storms and cascading failure indicators.

### 3) Product / Team View

- Feature adoption by workflow.
- Cost per successful outcome.
- Version comparison (prompt/workflow A vs B).
- Regression panel after releases.

### 4) Finance / Billing View

- Billable events, invoice preview, adjustments.
- Tenant-level chargeback/showback.
- Credits, discounts, and anomalies.
- Export-ready ledger.

### 5) Governance / Compliance View

- Policy violations and blocked actions.
- Audit timeline by user/agent.
- Data access + retention compliance checks.
- Human-in-the-loop approvals backlog.

## Alerting Ideas

- Spend spike: >X% over trailing 7-day baseline.
- Success rate drop below threshold for N minutes.
- Error type spike (timeouts/quota/auth) above baseline.
- P99 latency breach by provider/model.
- Budget threshold at 50/80/100%.
- Policy violation burst by workflow.

## Failure Taxonomy (Draft)

- Provider/API: `timeout`, `rate_limit`, `quota_exceeded`, `auth_failed`.
- Model behavior: `invalid_json`, `tool_selection_error`, `unsafe_output`.
- Tooling/runtime: `tool_timeout`, `tool_contract_mismatch`, `dependency_unavailable`.
- Workflow design: `loop_detected`, `missing_context`, `bad_routing`.
- Governance: `policy_blocked`, `approval_timeout`.

## Data Pipeline Sketch

- SDK/agent middleware emits structured events.
- Event bus (Kafka/PubSub/Kinesis) buffers ingestion.
- Stream processor computes aggregates + anomaly features.
- OLTP store for raw traces; OLAP warehouse for dashboard queries.
- Metrics backend for low-latency timeseries charts.
- Long-term archival for audits and forensics.

## Cost Controls

- Per-agent and per-tenant hard/soft budgets.
- Dynamic model routing (cheap model first, escalate on confidence).
- Context window trimming/summarization policies.
- Retry caps and circuit breakers.
- Caching policy for deterministic sub-tasks.
- Tool usage allowlist per workflow.

## Governance Controls

- RBAC: who can deploy prompts/workflows/policies.
- Signed prompt/workflow versions with approval workflow.
- Policy-as-code with test suites.
- Immutable audit logs for critical actions.
- Break-glass access with mandatory reason capture.

## MVP Scope (4-6 Weeks)

- Capture events for request/step/tool/failure/cost.
- Basic dashboard: spend, usage, success rate, top errors.
- Alerts for spend spike + success-rate drop.
- Drilldown to trace view per failed request.
- Tenant budget thresholds and monthly forecast.

## Phase 2

- Chargeback automation + invoice export.
- Anomaly detection (cost/latency/errors) with seasonality.
- A/B test analytics by prompt/workflow version.
- Policy simulation mode before enforcement.
- Quality scoring integrations and eval dashboards.

## Open Decisions

- Build vs buy for observability backend?
- Single-pane with existing APM vs standalone AgentOps UI?
- Real-time requirements: 5s, 30s, or 5min freshness?
- Billing source of truth: provider invoices vs internal metering?
- Multi-cloud / data residency constraints?

## Quick Next Steps

- Pick canonical event schema and required fields.
- Instrument one high-volume workflow first.
- Define SLOs: success rate, latency, and budget variance.
- Ship MVP dashboard + two high-signal alerts.
- Run weekly cost/failure review with owners.

## Codex Session Continuity

- Handoff files live in `notes/codex-sessions`.
- Start a new handoff: `scripts/codex-session start --title "..."`
- Close with explicit next-step context: `scripts/codex-session close --summary "..." --next "..."`
- New Codex sessions should read `notes/codex-sessions/CURRENT.md` first.

## Startup Requirements

- Git
- Go `1.25+` (for `services/ingest`)
- `curl` (for health checks)

## Local Startup (Current MVP Scaffold)

```bash
cd services/ingest
go run ./cmd/ingest
```

Health checks:

```bash
curl -sS http://localhost:8080/healthz
curl -sS http://localhost:8080/readyz
```

Config vars:

- `PORT` (default `8080`)
- `APP_ENV` (default `dev`)
- `LOG_LEVEL` (default `info`, one of `debug|info|warn|error`)
- `SHUTDOWN_TIMEOUT` (default `10s`)

Run tests:

```bash
cd services/ingest
go test ./...
```
