package main

import (
	"context"
	"errors"
	"github.com/example/policy-language-runtime/internal/config"
	"github.com/example/policy-language-runtime/internal/platform"
	"github.com/example/policy-language-runtime/internal/transport"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		l.Error("invalid configuration", "error", err)
		os.Exit(2)
	}
	svc := platform.NewService(platform.NewMemoryStore())
	h := transport.Timeout{Next: transport.New(svc, l).Handler(), Duration: cfg.ExecTimeout}
	srv := &http.Server{Addr: cfg.Addr, Handler: h}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		if e := srv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			l.Error("server failed", "error", e)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := shutdownContext(5 * time.Second)
	defer cancel()
	srv.Shutdown(shutdown)
}

func shutdownContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
