package validation

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateAcceptsValidRunStartedPayload(t *testing.T) {
	validator := mustLoadRepoSchemaValidator(t)

	payload := map[string]any{
		"event_version": "v0",
		"event_id":      "123e4567-e89b-12d3-a456-426614174000",
		"event_type":    "run.started",
		"occurred_at":   "2026-02-07T21:00:00Z",
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

	errList := validator.Validate(payload)
	if len(errList) != 0 {
		t.Fatalf("Validate() errors = %v, want none", errList)
	}
}

func TestValidateRejectsMissingRequiredField(t *testing.T) {
	validator := mustLoadRepoSchemaValidator(t)

	payload := map[string]any{
		"event_version": "v0",
		"event_id":      "123e4567-e89b-12d3-a456-426614174000",
		"event_type":    "run.started",
		"occurred_at":   "2026-02-07T21:00:00Z",
		"tenant": map[string]any{
			"tenant_id":  "tenant-1",
			"project_id": "project-1",
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

	errList := validator.Validate(payload)
	if len(errList) == 0 {
		t.Fatal("expected validation errors, got none")
	}

	if !containsPath(errList, "$.tenant.workspace_id") {
		t.Fatalf("expected missing path $.tenant.workspace_id, got %v", errList)
	}
}

func TestValidateRejectsInvalidEnumValue(t *testing.T) {
	validator := mustLoadRepoSchemaValidator(t)

	payload := map[string]any{
		"event_version": "v0",
		"event_id":      "123e4567-e89b-12d3-a456-426614174000",
		"event_type":    "run.started",
		"occurred_at":   "2026-02-07T21:00:00Z",
		"tenant": map[string]any{
			"tenant_id":    "tenant-1",
			"workspace_id": "workspace-1",
			"project_id":   "project-1",
		},
		"run": map[string]any{
			"run_id":      "run-1",
			"agent_id":    "agent-1",
			"workflow_id": "workflow-1",
			"status":      "wrong-status",
		},
		"trace": map[string]any{
			"trace_id": "trace-1",
			"span_id":  "span-1",
		},
	}

	errList := validator.Validate(payload)
	if len(errList) == 0 {
		t.Fatal("expected validation errors, got none")
	}

	if !containsPath(errList, "$.run.status") {
		t.Fatalf("expected enum path $.run.status, got %v", errList)
	}
}

func TestValidateAppliesConditionalRequirements(t *testing.T) {
	validator := mustLoadRepoSchemaValidator(t)

	payload := map[string]any{
		"event_version": "v0",
		"event_id":      "123e4567-e89b-12d3-a456-426614174000",
		"event_type":    "run.completed",
		"occurred_at":   "2026-02-07T21:00:00Z",
		"tenant": map[string]any{
			"tenant_id":    "tenant-1",
			"workspace_id": "workspace-1",
			"project_id":   "project-1",
		},
		"run": map[string]any{
			"run_id":      "run-1",
			"agent_id":    "agent-1",
			"workflow_id": "workflow-1",
			"status":      "success",
		},
		"trace": map[string]any{
			"trace_id": "trace-1",
			"span_id":  "span-1",
		},
	}

	errList := validator.Validate(payload)
	if len(errList) == 0 {
		t.Fatal("expected validation errors for missing conditional fields")
	}

	if !containsPath(errList, "$.resource_usage") {
		t.Fatalf("expected missing path $.resource_usage, got %v", errList)
	}
	if !containsPath(errList, "$.cost") {
		t.Fatalf("expected missing path $.cost, got %v", errList)
	}
}

func mustLoadRepoSchemaValidator(t *testing.T) *EventValidator {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}

	schemaPath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "../../../../packages/schemas/agent-event-v0.schema.json"))
	validator, err := NewEventValidator(schemaPath)
	if err != nil {
		t.Fatalf("NewEventValidator() error = %v", err)
	}

	return validator
}

func containsPath(errList []Error, target string) bool {
	for _, err := range errList {
		if strings.TrimSpace(err.Path) == target {
			return true
		}
	}
	return false
}
