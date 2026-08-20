package adapter

import (
	"fmt"
	"unicode/utf8"

	"github.com/example/policy-language-runtime/internal/lexer/domain"
)

type Limits struct{ MaxBytes, MaxTokens, MaxLine int }

func DefaultLimits() Limits        { return Limits{MaxBytes: 1 << 20, MaxTokens: 10000, MaxLine: 10000} }
func (l Limits) Accept(n int) bool { return n > 0 && n <= l.MaxBytes }

func (l Limits) ValidateSource(src string) error {
	if l.MaxBytes <= 0 || l.MaxTokens <= 0 || l.MaxLine <= 0 {
		return &domain.ScanError{Cause: domain.ErrLimit, Pos: domain.Position{Line: 1, Column: 1}, Text: "invalid lexer limits"}
	}
	if len(src) > l.MaxBytes {
		return &domain.ScanError{Cause: domain.ErrLimit, Pos: domain.Position{Line: 1, Column: 1}, Text: fmt.Sprintf("source exceeds %d bytes", l.MaxBytes)}
	}
	line := 1
	column := 0
	for len(src) > 0 {
		r, size := utf8.DecodeRuneInString(src)
		src = src[size:]
		if r == '\n' {
			line++
			column = 0
			continue
		}
		column++
		if column > l.MaxLine {
			return &domain.ScanError{Cause: domain.ErrLimit, Pos: domain.Position{Line: line, Column: column}, Text: fmt.Sprintf("line exceeds %d runes", l.MaxLine)}
		}
	}
	return nil
}
