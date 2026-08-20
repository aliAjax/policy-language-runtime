package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddlewareCancelsWork(t *testing.T) {
	seen := make(chan error, 1)
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		seen <- r.Context().Err()
	})
	parent, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(parent)
	done := make(chan struct{})
	go func() {
		Timeout{Next: next, Duration: time.Second}.ServeHTTP(httptest.NewRecorder(), req)
		close(done)
	}()
	cancel()
	select {
	case err := <-seen:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("want parent cancellation, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("middleware did not propagate parent cancellation")
	}
	<-done
}

func TestGRPCCompileStopsAfterCancel(t *testing.T) {
	entered := make(chan struct{})
	observed := make(chan error, 1)
	g := &GRPCServer{Compiler: func(ctx context.Context, _ string) (bool, error) {
		close(entered)
		<-ctx.Done()
		observed <- ctx.Err()
		return false, ctx.Err()
	}}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := g.Compile(ctx, "return true;"); done <- err }()
	<-entered
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("want context canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("compile ignored cancellation")
	}
	select {
	case err := <-observed:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("worker saw the wrong context error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("compiler worker did not receive cancellation")
	}
}
