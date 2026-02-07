package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/francisbulus/agent-ops/services/ingest/internal/persistence"
	"github.com/francisbulus/agent-ops/services/ingest/internal/validation"
)

const maxEventBodyBytes int64 = 1 << 20 // 1MB

// EventValidator validates event payloads.
type EventValidator interface {
	Validate(payload any) []validation.Error
}

// EventStore persists validated events.
type EventStore interface {
	InsertEvent(ctx context.Context, payload map[string]any) (bool, error)
	GetOverviewMetrics(ctx context.Context, filter persistence.OverviewFilter) (persistence.OverviewMetrics, error)
}

// NewHandler returns the ingest service HTTP handler tree.
func NewHandler(logger *slog.Logger, validator EventValidator, store EventStore) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST /v1/events", func(w http.ResponseWriter, r *http.Request) {
		handlePostEvents(w, r, validator, store)
	})
	mux.HandleFunc("GET /v1/metrics/overview", func(w http.ResponseWriter, r *http.Request) {
		handleGetMetricsOverview(w, r, store)
	})

	return requestLogger(logger, mux)
}

func handlePostEvents(w http.ResponseWriter, r *http.Request, validator EventValidator, store EventStore) {
	if validator == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "validator_not_configured"})
		return
	}
	if store == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "store_not_configured"})
		return
	}

	payload, err := decodeJSONBody(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":   "invalid_json",
			"message": err.Error(),
		})
		return
	}

	validationErrors := validator.Validate(payload)
	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation_failed",
			"errors": validationErrors,
		})
		return
	}

	payloadMap, ok := payload.(map[string]any)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":   "invalid_payload_type",
			"message": "request body must be a JSON object",
		})
		return
	}

	inserted, err := store.InsertEvent(r.Context(), payloadMap)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "persist_failed",
			"message": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":    "accepted",
		"persisted": inserted,
	})
}

func handleGetMetricsOverview(w http.ResponseWriter, r *http.Request, store EventStore) {
	if store == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "store_not_configured"})
		return
	}

	filter, err := parseOverviewFilter(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_query",
			"message": err.Error(),
		})
		return
	}

	overview, err := store.GetOverviewMetrics(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "metrics_query_failed",
			"message": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

func parseOverviewFilter(r *http.Request) (persistence.OverviewFilter, error) {
	var filter persistence.OverviewFilter
	query := r.URL.Query()

	filter.TenantID = query.Get("tenant_id")
	filter.WorkspaceID = query.Get("workspace_id")
	filter.ProjectID = query.Get("project_id")
	filter.AgentID = query.Get("agent_id")
	filter.WorkflowID = query.Get("workflow_id")

	windowHoursRaw := query.Get("window_hours")
	if windowHoursRaw == "" {
		filter.WindowHours = 24
		return filter, nil
	}

	windowHours, err := strconv.Atoi(windowHoursRaw)
	if err != nil {
		return filter, fmt.Errorf("window_hours must be an integer")
	}
	if windowHours < 1 || windowHours > 168 {
		return filter, fmt.Errorf("window_hours must be between 1 and 168")
	}
	filter.WindowHours = windowHours

	return filter, nil
}

func decodeJSONBody(body io.ReadCloser) (any, error) {
	defer body.Close()

	limited := io.LimitReader(body, maxEventBodyBytes)
	dec := json.NewDecoder(limited)
	dec.UseNumber()

	var payload any
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}

	var trailing any
	if err := dec.Decode(&trailing); err != nil {
		if errors.Is(err, io.EOF) {
			return payload, nil
		}
		return nil, errors.New("request body must contain a single JSON object")
	}

	return nil, errors.New("request body must contain a single JSON object")
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (s *statusRecorder) WriteHeader(statusCode int) {
	s.statusCode = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		logger.Info("http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.statusCode),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("trace_id", r.Header.Get("X-Trace-ID")),
		)
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
