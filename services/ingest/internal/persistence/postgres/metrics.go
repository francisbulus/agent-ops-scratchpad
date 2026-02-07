package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/francisbulus/agent-ops/services/ingest/internal/persistence"
)

const (
	defaultWindowHours = 24
	maxWindowHours     = 24 * 7
)

// GetOverviewMetrics returns aggregate usage/cost/reliability metrics for dashboard overview.
func (s *Store) GetOverviewMetrics(ctx context.Context, filter persistence.OverviewFilter) (persistence.OverviewMetrics, error) {
	var out persistence.OverviewMetrics

	if s == nil || s.db == nil || s.queryRow == nil {
		return out, errors.New("event store is not configured")
	}

	windowHours := filter.WindowHours
	if windowHours == 0 {
		windowHours = defaultWindowHours
	}
	if windowHours < 1 || windowHours > maxWindowHours {
		return out, fmt.Errorf("window_hours must be between 1 and %d", maxWindowHours)
	}

	windowEnd := time.Now().UTC()
	windowStart := windowEnd.Add(-time.Duration(windowHours) * time.Hour)

	query, args := buildOverviewQuery(windowStart, windowEnd, filter)
	row := s.queryRow(ctx, query, args...)

	var totalRuns int64
	var successfulRuns int64
	var failedRuns int64
	var totalCostUSD float64
	var avgLatencyMS float64

	if err := row.Scan(&totalRuns, &successfulRuns, &failedRuns, &totalCostUSD, &avgLatencyMS); err != nil {
		return out, fmt.Errorf("query metrics overview: %w", err)
	}

	successRate := 0.0
	if totalRuns > 0 {
		successRate = (float64(successfulRuns) / float64(totalRuns)) * 100
	}

	filter.WindowHours = windowHours
	out = persistence.OverviewMetrics{
		WindowStart:    windowStart,
		WindowEnd:      windowEnd,
		WindowHours:    windowHours,
		Filters:        filter,
		TotalRuns:      totalRuns,
		SuccessfulRuns: successfulRuns,
		FailedRuns:     failedRuns,
		SuccessRate:    successRate,
		TotalCostUSD:   totalCostUSD,
		AvgLatencyMS:   avgLatencyMS,
	}

	return out, nil
}

func buildOverviewQuery(windowStart time.Time, windowEnd time.Time, filter persistence.OverviewFilter) (string, []any) {
	var b strings.Builder
	args := make([]any, 0, 8)

	b.WriteString(`
SELECT
  COALESCE(SUM(CASE WHEN event_type IN ('run.completed', 'run.failed') THEN 1 ELSE 0 END), 0) AS total_runs,
  COALESCE(SUM(CASE WHEN event_type = 'run.completed' THEN 1 ELSE 0 END), 0) AS successful_runs,
  COALESCE(SUM(CASE WHEN event_type = 'run.failed' THEN 1 ELSE 0 END), 0) AS failed_runs,
  COALESCE(SUM(cost_usd), 0) AS total_cost_usd,
  COALESCE(
    AVG(
      CASE
        WHEN event_type IN ('run.completed', 'run.failed')
             AND payload->'run'->>'latency_ms' IS NOT NULL
        THEN (payload->'run'->>'latency_ms')::DOUBLE PRECISION
      END
    ),
    0
  ) AS avg_latency_ms
FROM agent_events
WHERE occurred_at >= $1 AND occurred_at <= $2`)
	args = append(args, windowStart, windowEnd)

	nextArg := 3
	appendFilter := func(column string, value string) {
		if value == "" {
			return
		}
		b.WriteString(fmt.Sprintf(" AND %s = $%d", column, nextArg))
		args = append(args, value)
		nextArg++
	}

	appendFilter("tenant_id", filter.TenantID)
	appendFilter("workspace_id", filter.WorkspaceID)
	appendFilter("project_id", filter.ProjectID)
	appendFilter("agent_id", filter.AgentID)
	appendFilter("workflow_id", filter.WorkflowID)

	return b.String(), args
}
