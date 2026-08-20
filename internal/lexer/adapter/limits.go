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
		return fmt.Errorf("invalid lexer limits: %v", domain.ErrLimit)
	}
	if len(src) > l.MaxBytes {
		return fmt.Errorf("source exceeds %d bytes: %v", l.MaxBytes, domain.ErrLimit)
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
			return fmt.Errorf("line %d exceeds %d runes: %v", line, l.MaxLine, domain.ErrLimit)
		}
	}
	return nil
}
