CREATE TABLE IF NOT EXISTS agent_events (
  id BIGSERIAL PRIMARY KEY,
  event_id UUID NOT NULL UNIQUE,
  event_version TEXT NOT NULL,
  event_type TEXT NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  tenant_id TEXT NOT NULL,
  workspace_id TEXT NOT NULL,
  project_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  agent_id TEXT NOT NULL,
  workflow_id TEXT NOT NULL,
  trace_id TEXT NOT NULL,
  span_id TEXT NOT NULL,
  parent_span_id TEXT NULL,
  run_status TEXT NULL,
  error_type TEXT NULL,
  total_tokens BIGINT NULL,
  cost_usd NUMERIC(18, 6) NULL,
  payload JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_agent_events_tenant_occurred_at
  ON agent_events (tenant_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_events_workspace_occurred_at
  ON agent_events (workspace_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_events_project_occurred_at
  ON agent_events (project_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_events_event_type_occurred_at
  ON agent_events (event_type, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_events_run_id
  ON agent_events (run_id);

CREATE INDEX IF NOT EXISTS idx_agent_events_trace_id
  ON agent_events (trace_id);
