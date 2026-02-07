package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"
)

type fakeServer struct {
	listenCalled   chan struct{}
	shutdownCalled chan struct{}
}

func newFakeServer() *fakeServer {
	return &fakeServer{
		listenCalled:   make(chan struct{}),
		shutdownCalled: make(chan struct{}),
	}
}

func (f *fakeServer) ListenAndServe() error {
	close(f.listenCalled)
	<-f.shutdownCalled
	return http.ErrServerClosed
}

func (f *fakeServer) Shutdown(context.Context) error {
	select {
	case <-f.shutdownCalled:
	default:
		close(f.shutdownCalled)
	}
	return nil
}

func TestRunGracefulShutdownOnSignal(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	sigCh := make(chan os.Signal, 1)
	done := make(chan error, 1)
	srv := newFakeServer()

	go func() {
		done <- runServer(context.Background(), logger, 2*time.Second, sigCh, srv)
	}()

	select {
	case <-srv.listenCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not start in time")
	}

	sigCh <- os.Interrupt

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run() did not shut down in time after signal")
	}
}

func TestRunServerReturnsListenError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	wantErr := errors.New("listen failed")

	srv := &errorServer{err: wantErr}
	err := runServer(context.Background(), logger, time.Second, nil, srv)
	if !errors.Is(err, wantErr) {
		t.Fatalf("runServer() error = %v, want %v", err, wantErr)
	}
}

type errorServer struct {
	err error
}

func (e *errorServer) ListenAndServe() error {
	return e.err
}

func (e *errorServer) Shutdown(context.Context) error {
	return nil
}
