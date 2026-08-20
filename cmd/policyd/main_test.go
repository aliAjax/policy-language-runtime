package main

import (
	"testing"
	"time"
)

func TestShutdownDeadlineIsBounded(t *testing.T) {
	ctx, cancel := shutdownContext(25 * time.Millisecond)
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("shutdown context has no deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > 50*time.Millisecond {
		t.Fatalf("unexpected shutdown deadline: %s", remaining)
	}
}
