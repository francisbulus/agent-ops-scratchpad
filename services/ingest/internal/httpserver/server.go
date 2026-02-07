package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/francisbulus/agent-ops/services/ingest/internal/validation"
)

const maxEventBodyBytes int64 = 1 << 20 // 1MB

// EventValidator validates event payloads.
type EventValidator interface {
	Validate(payload any) []validation.Error
}

// NewHandler returns the ingest service HTTP handler tree.
func NewHandler(logger *slog.Logger, validator EventValidator) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST /v1/events", func(w http.ResponseWriter, r *http.Request) {
		handlePostEvents(w, r, validator)
	})

	return requestLogger(logger, mux)
}

func handlePostEvents(w http.ResponseWriter, r *http.Request, validator EventValidator) {
	if validator == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "validator_not_configured"})
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

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
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
