package application_test

import (
	"encoding/json"
	"testing"

	irapp "github.com/example/policy-language-runtime/internal/ir/application"
	irdomain "github.com/example/policy-language-runtime/internal/ir/domain"
	optimizeradapter "github.com/example/policy-language-runtime/internal/optimizer/adapter"
	optimizerapp "github.com/example/policy-language-runtime/internal/optimizer/application"
	parserdomain "github.com/example/policy-language-runtime/internal/parser/domain"
)

func TestASTCloneDetachesNestedBlocks(t *testing.T) {
	original := &parserdomain.Program{Statements: []parserdomain.Stmt{parserdomain.If{
		Cond: parserdomain.Literal{Value: true},
		Then: &parserdomain.Block{Statements: []parserdomain.Stmt{parserdomain.Return{Value: parserdomain.Call{Name: "contains", Args: []parserdomain.Expr{parserdomain.Literal{Value: "a"}}}}}},
	}}}
	cloned := parserdomain.CloneProgram(original)
	clonedIf := cloned.Statements[0].(parserdomain.If)
	clonedCall := clonedIf.Then.Statements[0].(parserdomain.Return).Value.(parserdomain.Call)
	clonedCall.Args[0] = parserdomain.Literal{Value: "changed"}
	originalCall := original.Statements[0].(parserdomain.If).Then.Statements[0].(parserdomain.Return).Value.(parserdomain.Call)
	if originalCall.Args[0].(parserdomain.Literal).Value != "a" {
		t.Fatal("AST clone shares nested call arguments")
	}
}

func TestCompilerReturnsDetachedDependencies(t *testing.T) {
	compiler := irapp.New()
	firstProgram := &parserdomain.Program{Statements: []parserdomain.Stmt{parserdomain.Return{Value: parserdomain.Variable{Name: "age"}}}}
	first, err := compiler.Compile(firstProgram)
	if err != nil {
		t.Fatal(err)
	}
	secondProgram := &parserdomain.Program{Statements: []parserdomain.Stmt{parserdomain.Return{Value: parserdomain.Variable{Name: "country"}}}}
	second, err := compiler.Compile(secondProgram)
	if err != nil || second.Deps[0] != "country" || second.Code[0].Arg != "country" {
		t.Fatalf("unexpected second result: %#v (%v)", second, err)
	}
	if first.Deps[0] != "age" || first.Code[0].Arg != "age" {
		t.Fatalf("second compile rewrote first result: %#v", first)
	}
}

func TestOptimizerDoesNotMutateSourceCode(t *testing.T) {
	source := irdomain.Program{Code: []irdomain.Instr{
		{Op: irdomain.LoadConst, Value: float64(1)},
		{Op: irdomain.LoadConst, Value: float64(2)},
		{Op: irdomain.Binary, Arg: "+"},
		{Op: irdomain.Return},
	}}
	optimized := optimizerapp.New().Optimize(source)
	optimized.Code[0].Value = float64(99)
	if len(source.Code) != 4 || source.Code[0].Value != float64(1) {
		t.Fatalf("optimizer mutated source: %#v", source)
	}
}

func TestSerializedProgramSurvivesOptimization(t *testing.T) {
	source := irdomain.Program{Code: []irdomain.Instr{{Op: irdomain.LoadConst, Value: float64(1)}}, Deps: []string{"tenant"}}
	before, _ := json.Marshal(source)
	optimized := optimizerapp.New().Optimize(source)
	optimized.Deps[0] = "changed"
	after, _ := json.Marshal(source)
	if string(before) != string(after) {
		t.Fatalf("serialized source changed: before=%s after=%s", before, after)
	}
}

func TestOptimizationReportOwnsDependencies(t *testing.T) {
	source := optimizeradapter.Report{Dependencies: 2, DependencyNames: []string{"region", "age"}}
	first := source.Clone()
	first.DependencyNames[0] = "changed"
	if source.DependencyNames[0] != "region" {
		t.Fatalf("report clone leaked dependency slice: %#v", source)
	}
}
