package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const insertEventSQL = `
INSERT INTO agent_events (
  event_id,
  event_version,
  event_type,
  occurred_at,
  tenant_id,
  workspace_id,
  project_id,
  run_id,
  agent_id,
  workflow_id,
  trace_id,
  span_id,
  parent_span_id,
  run_status,
  error_type,
  total_tokens,
  cost_usd,
  payload
)
VALUES (
  $1, $2, $3, $4,
  $5, $6, $7,
  $8, $9, $10,
  $11, $12, $13,
  $14, $15, $16, $17, $18
)
ON CONFLICT (event_id) DO NOTHING
`

type dbAPI interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}

// Store persists validated events in Postgres.
type Store struct {
	db dbAPI
}

// NewStore constructs a postgres-backed event store and verifies connectivity.
func NewStore(databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Store{db: db}, nil
}

// InsertEvent writes one validated event. It returns inserted=false for idempotent duplicates.
func (s *Store) InsertEvent(ctx context.Context, payload map[string]any) (bool, error) {
	if s == nil || s.db == nil {
		return false, errors.New("event store is not configured")
	}

	row, err := buildEventRow(payload)
	if err != nil {
		return false, err
	}

	result, err := s.db.ExecContext(ctx, insertEventSQL,
		row.EventID,
		row.EventVersion,
		row.EventType,
		row.OccurredAt,
		row.TenantID,
		row.WorkspaceID,
		row.ProjectID,
		row.RunID,
		row.AgentID,
		row.WorkflowID,
		row.TraceID,
		row.SpanID,
		row.ParentSpanID,
		row.RunStatus,
		row.ErrorType,
		row.TotalTokens,
		row.CostUSD,
		row.Payload,
	)
	if err != nil {
		return false, fmt.Errorf("insert agent event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read insert rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// Ready performs a lightweight database readiness check.
func (s *Store) Ready(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("event store is not configured")
	}
	return s.db.PingContext(ctx)
}

// Close releases database resources.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

type eventRow struct {
	EventID      string
	EventVersion string
	EventType    string
	OccurredAt   time.Time
	TenantID     string
	WorkspaceID  string
	ProjectID    string
	RunID        string
	AgentID      string
	WorkflowID   string
	TraceID      string
	SpanID       string
	ParentSpanID *string
	RunStatus    *string
	ErrorType    *string
	TotalTokens  *int64
	CostUSD      *float64
	Payload      []byte
}

func buildEventRow(payload map[string]any) (eventRow, error) {
	var row eventRow

	eventID, err := requireString(payload, "event_id")
	if err != nil {
		return row, err
	}
	eventVersion, err := requireString(payload, "event_version")
	if err != nil {
		return row, err
	}
	eventType, err := requireString(payload, "event_type")
	if err != nil {
		return row, err
	}
	occurredAtString, err := requireString(payload, "occurred_at")
	if err != nil {
		return row, err
	}
	occurredAt, err := time.Parse(time.RFC3339, occurredAtString)
	if err != nil {
		return row, fmt.Errorf("$.occurred_at must be RFC3339 date-time")
	}

	tenantID, err := requireString(payload, "tenant", "tenant_id")
	if err != nil {
		return row, err
	}
	workspaceID, err := requireString(payload, "tenant", "workspace_id")
	if err != nil {
		return row, err
	}
	projectID, err := requireString(payload, "tenant", "project_id")
	if err != nil {
		return row, err
	}

	runID, err := requireString(payload, "run", "run_id")
	if err != nil {
		return row, err
	}
	agentID, err := requireString(payload, "run", "agent_id")
	if err != nil {
		return row, err
	}
	workflowID, err := requireString(payload, "run", "workflow_id")
	if err != nil {
		return row, err
	}

	traceID, err := requireString(payload, "trace", "trace_id")
	if err != nil {
		return row, err
	}
	spanID, err := requireString(payload, "trace", "span_id")
	if err != nil {
		return row, err
	}

	parentSpanID, err := optionalString(payload, "trace", "parent_span_id")
	if err != nil {
		return row, err
	}
	runStatus, err := optionalString(payload, "run", "status")
	if err != nil {
		return row, err
	}
	errorType, err := optionalString(payload, "error", "error_type")
	if err != nil {
		return row, err
	}
	totalTokens, err := optionalInt64(payload, "resource_usage", "total_tokens")
	if err != nil {
		return row, err
	}
	costUSD, err := optionalFloat64(payload, "cost", "cost_usd")
	if err != nil {
		return row, err
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return row, fmt.Errorf("marshal payload json: %w", err)
	}

	row = eventRow{
		EventID:      eventID,
		EventVersion: eventVersion,
		EventType:    eventType,
		OccurredAt:   occurredAt,
		TenantID:     tenantID,
		WorkspaceID:  workspaceID,
		ProjectID:    projectID,
		RunID:        runID,
		AgentID:      agentID,
		WorkflowID:   workflowID,
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		RunStatus:    runStatus,
		ErrorType:    errorType,
		TotalTokens:  totalTokens,
		CostUSD:      costUSD,
		Payload:      rawPayload,
	}

	return row, nil
}

func requireString(root map[string]any, path ...string) (string, error) {
	value, ok, err := lookup(root, path...)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("%s is required", joinPath(path))
	}

	str, ok := value.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("%s must be a non-empty string", joinPath(path))
	}

	return str, nil
}

func optionalString(root map[string]any, path ...string) (*string, error) {
	value, ok, err := lookup(root, path...)
	if err != nil {
		return nil, err
	}
	if !ok || value == nil {
		return nil, nil
	}

	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("%s must be a string", joinPath(path))
	}

	return &str, nil
}

func optionalInt64(root map[string]any, path ...string) (*int64, error) {
	value, ok, err := lookup(root, path...)
	if err != nil {
		return nil, err
	}
	if !ok || value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil, fmt.Errorf("%s must be an integer", joinPath(path))
		}
		return &i, nil
	case float64:
		i := int64(v)
		if float64(i) != v {
			return nil, fmt.Errorf("%s must be an integer", joinPath(path))
		}
		return &i, nil
	case int:
		i := int64(v)
		return &i, nil
	case int64:
		return &v, nil
	default:
		return nil, fmt.Errorf("%s must be an integer", joinPath(path))
	}
}

func optionalFloat64(root map[string]any, path ...string) (*float64, error) {
	value, ok, err := lookup(root, path...)
	if err != nil {
		return nil, err
	}
	if !ok || value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return nil, fmt.Errorf("%s must be a number", joinPath(path))
		}
		return &f, nil
	case float64:
		return &v, nil
	case int:
		f := float64(v)
		return &f, nil
	default:
		return nil, fmt.Errorf("%s must be a number", joinPath(path))
	}
}

func lookup(root map[string]any, path ...string) (any, bool, error) {
	if len(path) == 0 {
		return nil, false, errors.New("path is required")
	}

	current := any(root)
	for idx, key := range path {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("%s must be an object", joinPath(path[:idx]))
		}

		value, exists := obj[key]
		if !exists {
			return nil, false, nil
		}

		current = value
	}

	return current, true, nil
}

func joinPath(path []string) string {
	return "$." + join(path, ".")
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += sep + parts[i]
	}
	return out
}
