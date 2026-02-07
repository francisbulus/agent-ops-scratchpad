package persistence

import "time"

// OverviewFilter defines query constraints for metrics overview reads.
type OverviewFilter struct {
	WindowHours int    `json:"window_hours"`
	TenantID    string `json:"tenant_id,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	AgentID     string `json:"agent_id,omitempty"`
	WorkflowID  string `json:"workflow_id,omitempty"`
}

// OverviewMetrics is the deterministic response contract for GET /v1/metrics/overview.
type OverviewMetrics struct {
	WindowStart    time.Time      `json:"window_start"`
	WindowEnd      time.Time      `json:"window_end"`
	WindowHours    int            `json:"window_hours"`
	Filters        OverviewFilter `json:"filters"`
	TotalRuns      int64          `json:"total_runs"`
	SuccessfulRuns int64          `json:"successful_runs"`
	FailedRuns     int64          `json:"failed_runs"`
	SuccessRate    float64        `json:"success_rate"`
	TotalCostUSD   float64        `json:"total_cost_usd"`
	AvgLatencyMS   float64        `json:"avg_latency_ms"`
}
