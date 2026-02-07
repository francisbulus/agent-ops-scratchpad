package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type stubStore struct {
	inserted bool
	err      error
}

func (s stubStore) InsertEvent(context.Context, map[string]any) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.inserted, nil
}

func TestHealthAndReadyEndpoints(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{}, stubStore{inserted: true})

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
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{}, stubStore{inserted: true})

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
	if body["persisted"] != true {
		t.Fatalf("persisted = %v, want true", body["persisted"])
	}
}

func TestPostEventsInvalidJSON(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{}, stubStore{inserted: true})

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
	}}, stubStore{inserted: true})

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

func TestPostEventsPersistFailure(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), stubValidator{}, stubStore{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewBufferString(`{"event_id":"x"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["error"] != "persist_failed" {
		t.Fatalf("error = %v, want persist_failed", body["error"])
	}
}
