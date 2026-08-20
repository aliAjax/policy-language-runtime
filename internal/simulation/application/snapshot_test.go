package application_test

import (
	"context"
	"testing"

	explainapp "github.com/example/policy-language-runtime/internal/explain/application"
	irdomain "github.com/example/policy-language-runtime/internal/ir/domain"
	runtimeadapter "github.com/example/policy-language-runtime/internal/runtime/adapter"
	runtimeapp "github.com/example/policy-language-runtime/internal/runtime/application"
	runtimedomain "github.com/example/policy-language-runtime/internal/runtime/domain"
	simulationapp "github.com/example/policy-language-runtime/internal/simulation/application"
	"github.com/example/policy-language-runtime/internal/simulation/domain"
)

func decisionProgram(value bool) irdomain.Program {
	return irdomain.Program{Code: []irdomain.Instr{{Op: irdomain.LoadConst, Value: value}, {Op: irdomain.Return}}}
}

func TestSimulationReportKeepsOriginalSamples(t *testing.T) {
	input := map[string]any{"roles": []string{"reader"}}
	samples := []domain.Sample{{ID: "s1", Input: input}}
	report := simulationapp.New(runtimeapp.New()).Compare(context.Background(), decisionProgram(true), decisionProgram(false), samples)
	input["roles"].([]string)[0] = "admin"
	samples[0].Input["new"] = true
	if got := report.Samples[0].Input["roles"].([]string)[0]; got != "reader" {
		t.Fatalf("report sample changed to %q", got)
	}
	if _, ok := report.Samples[0].Input["new"]; ok {
		t.Fatal("report shares the caller input map")
	}
}

func TestExplanationChildrenAreDetached(t *testing.T) {
	result := runtimedomain.Result{Decision: runtimedomain.Allow, Steps: []string{"load", "return"}, Value: []string{"allow"}}
	explanation := explainapp.New().Build(result)
	result.Steps[0] = "changed"
	result.Value.([]string)[0] = "deny"
	if explanation.Root.Children[0].Value != "load" || explanation.Root.Value.([]string)[0] != "allow" {
		t.Fatalf("explanation shares result storage: %#v", explanation)
	}
}

func TestSecondComparisonCannotRewriteFirst(t *testing.T) {
	svc := simulationapp.New(runtimeapp.New())
	first := svc.Compare(context.Background(), decisionProgram(true), decisionProgram(false), []domain.Sample{{ID: "s1", Input: map[string]any{}}})
	second := svc.Compare(context.Background(), decisionProgram(false), decisionProgram(true), []domain.Sample{{ID: "s2", Input: map[string]any{}}})
	if len(first.Changes) != 1 || len(second.Changes) != 1 {
		t.Fatalf("comparison reports share state: first=%#v second=%#v", first, second)
	}
	if first.Changes[0].SampleID != "s1" || second.Changes[0].SampleID != "s2" {
		t.Fatalf("second comparison rewrote first: first=%#v second=%#v", first, second)
	}
}

func TestBudgetSnapshotOwnsTrace(t *testing.T) {
	budget := runtimeadapter.NewBudget(0, 10)
	budget.Record("load")
	snapshot := budget.Snapshot()
	snapshot.Trace[0] = "changed"
	if budget.Trace[0] != "load" {
		t.Fatalf("budget trace escaped through snapshot: %#v", budget)
	}
}
