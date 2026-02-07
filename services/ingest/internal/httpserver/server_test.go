package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthAndReadyEndpoints(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)))

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
