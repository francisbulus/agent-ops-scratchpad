package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/francisbulus/agent-ops/services/ingest/internal/persistence"
)

type fakeScanRow struct {
	values []any
	err    error
}

func (f fakeScanRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}
	if len(f.values) != len(dest) {
		return errors.New("mismatched scan values")
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *int64:
			*d = f.values[i].(int64)
		case *float64:
			*d = f.values[i].(float64)
		default:
			return errors.New("unsupported scan destination")
		}
	}
	return nil
}

func TestBuildOverviewQueryIncludesFilters(t *testing.T) {
	start := time.Date(2026, 2, 7, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	query, args := buildOverviewQuery(start, end, persistence.OverviewFilter{
		TenantID:    "tenant-1",
		WorkspaceID: "workspace-1",
		ProjectID:   "project-1",
		AgentID:     "agent-1",
		WorkflowID:  "workflow-1",
	})

	if !strings.Contains(query, "tenant_id = $3") {
		t.Fatalf("query missing tenant filter: %s", query)
	}
	if !strings.Contains(query, "workspace_id = $4") {
		t.Fatalf("query missing workspace filter: %s", query)
	}
	if !strings.Contains(query, "project_id = $5") {
		t.Fatalf("query missing project filter: %s", query)
	}
	if !strings.Contains(query, "agent_id = $6") {
		t.Fatalf("query missing agent filter: %s", query)
	}
	if !strings.Contains(query, "workflow_id = $7") {
		t.Fatalf("query missing workflow filter: %s", query)
	}

	if len(args) != 7 {
		t.Fatalf("args len = %d, want 7", len(args))
	}
}

func TestGetOverviewMetricsReturnsComputedValues(t *testing.T) {
	store := &Store{
		db: &fakeDB{},
		queryRow: func(_ context.Context, _ string, _ ...any) rowScanner {
			return fakeScanRow{values: []any{
				int64(10),
				int64(8),
				int64(2),
				float64(12.34),
				float64(150),
			}}
		},
	}

	overview, err := store.GetOverviewMetrics(context.Background(), persistence.OverviewFilter{WindowHours: 24})
	if err != nil {
		t.Fatalf("GetOverviewMetrics() error = %v", err)
	}

	if overview.TotalRuns != 10 {
		t.Fatalf("TotalRuns = %d, want 10", overview.TotalRuns)
	}
	if overview.SuccessfulRuns != 8 {
		t.Fatalf("SuccessfulRuns = %d, want 8", overview.SuccessfulRuns)
	}
	if overview.FailedRuns != 2 {
		t.Fatalf("FailedRuns = %d, want 2", overview.FailedRuns)
	}
	if overview.SuccessRate != 80 {
		t.Fatalf("SuccessRate = %v, want 80", overview.SuccessRate)
	}
	if overview.TotalCostUSD != 12.34 {
		t.Fatalf("TotalCostUSD = %v, want 12.34", overview.TotalCostUSD)
	}
	if overview.AvgLatencyMS != 150 {
		t.Fatalf("AvgLatencyMS = %v, want 150", overview.AvgLatencyMS)
	}
}

func TestGetOverviewMetricsEmptyDataset(t *testing.T) {
	store := &Store{
		db: &fakeDB{},
		queryRow: func(_ context.Context, _ string, _ ...any) rowScanner {
			return fakeScanRow{values: []any{
				int64(0),
				int64(0),
				int64(0),
				float64(0),
				float64(0),
			}}
		},
	}

	overview, err := store.GetOverviewMetrics(context.Background(), persistence.OverviewFilter{})
	if err != nil {
		t.Fatalf("GetOverviewMetrics() error = %v", err)
	}

	if overview.TotalRuns != 0 || overview.SuccessRate != 0 || overview.TotalCostUSD != 0 {
		t.Fatalf("unexpected overview values: %+v", overview)
	}
	if overview.WindowHours != 24 {
		t.Fatalf("WindowHours = %d, want 24", overview.WindowHours)
	}
}

func TestGetOverviewMetricsRejectsBadWindow(t *testing.T) {
	store := &Store{
		db: &fakeDB{},
		queryRow: func(_ context.Context, _ string, _ ...any) rowScanner {
			return fakeScanRow{values: []any{
				int64(0),
				int64(0),
				int64(0),
				float64(0),
				float64(0),
			}}
		},
	}

	_, err := store.GetOverviewMetrics(context.Background(), persistence.OverviewFilter{WindowHours: 0})
	if err != nil {
		t.Fatalf("window_hours 0 should default, error = %v", err)
	}

	_, err = store.GetOverviewMetrics(context.Background(), persistence.OverviewFilter{WindowHours: 999})
	if err == nil {
		t.Fatal("expected error for invalid window hours")
	}
}
