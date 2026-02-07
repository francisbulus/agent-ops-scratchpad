package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francisbulus/agent-ops/services/ingest/internal/validation"
)

type stubValidator struct {
	errList []validation.Error
}

func (s stubValidator) Validate(any) []validation.Error {
	return s.errList
}

func TestHealthAndReadyEndpoints(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{})

	tests := []string{"/healthz", "/readyz"}
	for _, path := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, rr.Code, http.StatusOK)
		}

		if rr.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("%s content-type = %q, want application/json", path, rr.Header().Get("Content-Type"))
		}
	}
}

func TestPostEventsAccepted(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{})

	req := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewBufferString(`{"event_id":"x"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["status"] != "accepted" {
		t.Fatalf("status body = %v, want accepted", body["status"])
	}
}

func TestPostEventsInvalidJSON(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{})

	req := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewBufferString(`{"broken":`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["error"] != "invalid_json" {
		t.Fatalf("error = %v, want invalid_json", body["error"])
	}
}

func TestPostEventsValidationErrors(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{errList: []validation.Error{
		{Path: "$.event_type", Message: "must be one of allowed enum values"},
	}})

	req := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewBufferString(`{"event_type":"bad"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["error"] != "validation_failed" {
		t.Fatalf("error = %v, want validation_failed", body["error"])
	}

	errItems, ok := body["errors"].([]any)
	if !ok || len(errItems) == 0 {
		t.Fatalf("errors = %v, want non-empty array", body["errors"])
	}

	firstErr, ok := errItems[0].(map[string]any)
	if !ok {
		t.Fatalf("first error = %T, want object", errItems[0])
	}
	if firstErr["path"] != "$.event_type" {
		t.Fatalf("error path = %v, want $.event_type", firstErr["path"])
	}
}
