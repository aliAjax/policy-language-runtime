package application_test

import (
	"errors"
	"testing"

	lexeradapter "github.com/example/policy-language-runtime/internal/lexer/adapter"
	lexerapp "github.com/example/policy-language-runtime/internal/lexer/application"
	lexerdomain "github.com/example/policy-language-runtime/internal/lexer/domain"
	parserapp "github.com/example/policy-language-runtime/internal/parser/application"
)

func TestLexerLimitErrorChain(t *testing.T) {
	limits := lexeradapter.DefaultLimits()
	limits.MaxLine = 3
	_, err := lexerapp.NewWithLimits("return true;", limits).All()
	if !errors.Is(err, lexerdomain.ErrLimit) {
		t.Fatalf("want limit sentinel, got %v", err)
	}
}

func TestUnexpectedRuneKeepsPosition(t *testing.T) {
	_, err := lexerapp.New("\n  @", 10).All()
	var scan *lexerdomain.ScanError
	if !errors.As(err, &scan) || !errors.Is(err, lexerdomain.ErrUnexpectedRune) {
		t.Fatalf("want unexpected-rune scan error, got %v", err)
	}
	if scan.Pos.Line != 2 || scan.Pos.Column != 3 {
		t.Fatalf("want 2:3, got %d:%d", scan.Pos.Line, scan.Pos.Column)
	}
}

func TestUnterminatedStringClassification(t *testing.T) {
	_, err := lexerapp.New(`return "open`, 10).All()
	if !errors.Is(err, lexerdomain.ErrUnterminatedString) {
		t.Fatalf("want unterminated string sentinel, got %v", err)
	}
}

func TestParserDiagnosticUnwrapsLexerFailure(t *testing.T) {
	limits := lexeradapter.DefaultLimits()
	_, err := parserapp.ParseSource(`return "open`, limits)
	if !errors.Is(err, lexerdomain.ErrUnterminatedString) {
		t.Fatalf("parser diagnostic lost lexer error: %v", err)
	}
}
