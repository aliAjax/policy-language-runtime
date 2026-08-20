package config

import (
	"testing"
	"time"
)

func TestConfiguredDeadlineReachesHandler(t *testing.T) {
	t.Setenv("POLICY_CONFIG", t.TempDir()+"/missing.yaml")
	t.Setenv("POLICY_EXEC_TIMEOUT", "37ms")
	got := Load()
	if got.ExecTimeout != 37*time.Millisecond {
		t.Fatalf("want 37ms, got %s", got.ExecTimeout)
	}
}
