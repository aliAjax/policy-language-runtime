package application_test

import (
	"strings"
	"testing"

	builtinapp "github.com/example/policy-language-runtime/internal/builtin/application"
	builtindomain "github.com/example/policy-language-runtime/internal/builtin/domain"
	parserdomain "github.com/example/policy-language-runtime/internal/parser/domain"
	typecheckadapter "github.com/example/policy-language-runtime/internal/typecheck/adapter"
	typecheckapp "github.com/example/policy-language-runtime/internal/typecheck/application"
	typecheckdomain "github.com/example/policy-language-runtime/internal/typecheck/domain"
)

func callProgram(name string) *parserdomain.Program {
	return &parserdomain.Program{Statements: []parserdomain.Stmt{parserdomain.Return{Value: parserdomain.Call{Name: name}}}}
}

func hasCode(ds []typecheckdomain.Diagnostic, code string) bool {
	for _, d := range ds {
		if d.Code == code {
			return true
		}
	}
	return false
}

func TestTypedNilRegistryRejected(t *testing.T) {
	var concrete *builtinapp.Registry
	var registry builtindomain.Registry = concrete
	ds := typecheckapp.NewWithRegistry(registry).Check(callProgram("custom"))
	if !hasCode(ds, "registry_unavailable") {
		t.Fatalf("typed nil registry was not rejected: %#v", ds)
	}
}

func TestZeroValueCheckerReturnsDiagnostic(t *testing.T) {
	var checker typecheckapp.Checker
	p := &parserdomain.Program{Statements: []parserdomain.Stmt{parserdomain.Return{Value: parserdomain.Unary{Op: "-", Value: parserdomain.Literal{Value: true}}}}}
	ds := checker.Check(p)
	if len(ds) == 0 || !ds[0].Valid() {
		t.Fatalf("zero value checker did not return a diagnostic: %#v", ds)
	}
}

func TestUnknownBuiltinIsNotSilentlyAccepted(t *testing.T) {
	registry := builtinapp.New()
	ds := typecheckapp.NewWithRegistry(registry).Check(callProgram("geofence"))
	if !hasCode(ds, "unknown_builtin") {
		t.Fatalf("unknown builtin was accepted: %#v", ds)
	}
}

func TestNilDiagnosticFormatsSafely(t *testing.T) {
	if got := typecheckadapter.Format(nil); got != "" {
		t.Fatalf("nil diagnostics should format empty, got %q", got)
	}
	got := typecheckadapter.Format([]typecheckdomain.Diagnostic{{}, {Code: "bad", Message: "broken", Severity: "error"}})
	if strings.Contains(got, ": \n") || !strings.Contains(got, "error[bad]: broken") {
		t.Fatalf("unexpected diagnostic output %q", got)
	}
}
