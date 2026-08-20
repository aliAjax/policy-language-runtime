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
	Timeout{Next: next, Duration: 10 * time.Millisecond}.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if err := <-seen; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want deadline exceeded, got %v", err)
	}
}

func TestGRPCCompileStopsAfterCancel(t *testing.T) {
	entered := make(chan struct{})
	g := &GRPCServer{Compiler: func(ctx context.Context, _ string) (bool, error) {
		close(entered)
		<-ctx.Done()
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
}
