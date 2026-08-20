package application_test

import (
	"context"
	"sync"
	"testing"

	irdomain "github.com/example/policy-language-runtime/internal/ir/domain"
	runtimeapp "github.com/example/policy-language-runtime/internal/runtime/application"
	runtimedomain "github.com/example/policy-language-runtime/internal/runtime/domain"
	"github.com/example/policy-language-runtime/internal/runtime/infrastructure"
)

func TestEvaluatorRegisterConcurrentWithEval(t *testing.T) {
	e := runtimeapp.New()
	program := irdomain.Program{Code: []irdomain.Instr{
		{Op: irdomain.LoadConst, Value: true},
		{Op: irdomain.Call, Arg: "flip:1"},
		{Op: irdomain.Return},
	}}
	e.Register("flip", func(_ context.Context, args []any) (any, error) { return args[0], nil })
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 500; i++ {
			e.Register("flip", func(_ context.Context, args []any) (any, error) { return args[0], nil })
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 500; i++ {
			if _, err := e.Eval(context.Background(), program, map[string]any{}); err != nil {
				t.Errorf("eval failed: %v", err)
				return
			}
		}
	}()
	close(start)
	wg.Wait()
}

func TestCacheReturnsDetachedProgram(t *testing.T) {
	c := infrastructure.New()
	original := irdomain.Program{Code: []irdomain.Instr{{Op: irdomain.LoadConst, Value: true}}, Deps: []string{"age"}}
	c.Put("p", original)
	original.Code[0].Value = false
	original.Deps[0] = "changed"
	first, _ := c.Get("p")
	first.Code[0].Value = false
	first.Deps[0] = "mutated"
	second, _ := c.Get("p")
	if second.Code[0].Value != true || second.Deps[0] != "age" {
		t.Fatalf("cache leaked mutable program: %#v", second)
	}
}

func TestMetricsSnapshotConcurrentUpdates(t *testing.T) {
	var m infrastructure.Metrics
	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 0; worker < 2; worker++ {
		wg.Add(1)
		go func(failed bool) {
			defer wg.Done()
			<-start
			for i := 0; i < 1000; i++ {
				m.Record(failed)
				requests, errorsCount := m.Snapshot()
				if errorsCount > requests {
					t.Errorf("impossible snapshot: requests=%d errors=%d", requests, errorsCount)
					return
				}
			}
		}(worker == 1)
	}
	close(start)
	wg.Wait()
	requests, errorsCount := m.Snapshot()
	if requests != 2000 || errorsCount != 1000 {
		t.Fatalf("unexpected totals: requests=%d errors=%d", requests, errorsCount)
	}
}

func TestResultCloneDetachesSteps(t *testing.T) {
	original := runtimedomain.Result{Steps: []string{"load", "compare"}, Value: []string{"allow"}}
	cloned := original.Clone()
	cloned.Steps[0] = "changed"
	cloned.Value.([]string)[0] = "deny"
	if original.Steps[0] != "load" || original.Value.([]string)[0] != "allow" {
		t.Fatalf("clone mutated source: %#v", original)
	}
}
