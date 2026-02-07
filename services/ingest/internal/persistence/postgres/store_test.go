package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeResult struct {
	rows int64
}

func (f fakeResult) LastInsertId() (int64, error) {
	return 0, errors.New("unsupported")
}

func (f fakeResult) RowsAffected() (int64, error) {
	return f.rows, nil
}

type fakeDB struct {
	execErr  error
	result   sql.Result
	query    string
	args     []any
	pingErr  error
	closeErr error
}

func (f *fakeDB) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.query = query
	f.args = args
	if f.execErr != nil {
		return nil, f.execErr
	}
	if f.result != nil {
		return f.result, nil
	}
	return fakeResult{rows: 1}, nil
}

func (f *fakeDB) PingContext(context.Context) error {
	return f.pingErr
}

func (f *fakeDB) Close() error {
	return f.closeErr
}

func TestInsertEventPersistsRow(t *testing.T) {
	db := &fakeDB{result: fakeResult{rows: 1}}
	store := &Store{db: db}

	inserted, err := store.InsertEvent(context.Background(), validPayload())
	if err != nil {
		t.Fatalf("InsertEvent() error = %v", err)
	}
	if !inserted {
		t.Fatal("InsertEvent() inserted = false, want true")
	}
	if !strings.Contains(db.query, "INSERT INTO agent_events") {
		t.Fatalf("query = %q, want INSERT statement", db.query)
	}
	if got := len(db.args); got != 18 {
		t.Fatalf("args len = %d, want 18", got)
	}
}

func TestInsertEventDuplicateIsIdempotent(t *testing.T) {
	db := &fakeDB{result: fakeResult{rows: 0}}
	store := &Store{db: db}

	inserted, err := store.InsertEvent(context.Background(), validPayload())
	if err != nil {
		t.Fatalf("InsertEvent() error = %v", err)
	}
	if inserted {
		t.Fatal("InsertEvent() inserted = true, want false for duplicate")
	}
}

func TestInsertEventWriteError(t *testing.T) {
	db := &fakeDB{execErr: errors.New("write failed")}
	store := &Store{db: db}

	_, err := store.InsertEvent(context.Background(), validPayload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "insert agent event") {
		t.Fatalf("error = %v, want insert context", err)
	}
}

func TestInsertEventMissingRequiredField(t *testing.T) {
	db := &fakeDB{}
	store := &Store{db: db}

	payload := validPayload()
	tenant := payload["tenant"].(map[string]any)
	delete(tenant, "workspace_id")

	_, err := store.InsertEvent(context.Background(), payload)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
	if !strings.Contains(err.Error(), "$.tenant.workspace_id") {
		t.Fatalf("error = %v, want missing path", err)
	}
}

func TestReadyAndClose(t *testing.T) {
	db := &fakeDB{}
	store := &Store{db: db}

	if err := store.Ready(context.Background()); err != nil {
		t.Fatalf("Ready() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func validPayload() map[string]any {
	return map[string]any{
		"event_version": "v0",
		"event_id":      "550e8400-e29b-41d4-a716-446655440000",
		"event_type":    "run.started",
		"occurred_at":   time.Now().UTC().Format(time.RFC3339),
		"tenant": map[string]any{
			"tenant_id":    "tenant-1",
			"workspace_id": "workspace-1",
			"project_id":   "project-1",
		},
		"run": map[string]any{
			"run_id":      "run-1",
			"agent_id":    "agent-1",
			"workflow_id": "workflow-1",
			"status":      "started",
		},
		"trace": map[string]any{
			"trace_id": "trace-1",
			"span_id":  "span-1",
		},
	}
}
