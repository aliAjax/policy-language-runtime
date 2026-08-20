package application

import (
	"fmt"
	"unicode"

	"github.com/example/policy-language-runtime/internal/lexer/adapter"
	"github.com/example/policy-language-runtime/internal/lexer/domain"
)

type Lexer struct {
	src                  []rune
	offset, line, column int
	maxTokens            int
	limits               adapter.Limits
	validationErr        error
}

func New(src string, maxTokens int) *Lexer {
	if maxTokens <= 0 {
		maxTokens = 10000
	}
	limits := adapter.DefaultLimits()
	limits.MaxTokens = maxTokens
	return NewWithLimits(src, limits)
}

func NewWithLimits(src string, limits adapter.Limits) *Lexer {
	return &Lexer{src: []rune(src), line: 1, column: 1, maxTokens: limits.MaxTokens, limits: limits, validationErr: limits.ValidateSource(src)}
}

func (l *Lexer) Next() (domain.Token, error) {
	if l.validationErr != nil {
		return domain.Token{}, l.validationErr
	}
	for l.offset < len(l.src) && unicode.IsSpace(l.src[l.offset]) {
		l.advance()
	}
	if l.offset >= len(l.src) {
		return l.tok(domain.EOF, ""), nil
	}
	start := l.pos()
	r := l.src[l.offset]
	if unicode.IsLetter(r) || r == '_' {
		var b []rune
		for l.offset < len(l.src) && (unicode.IsLetter(l.src[l.offset]) || unicode.IsDigit(l.src[l.offset]) || l.src[l.offset] == '_') {
			b = append(b, l.src[l.offset])
			l.advance()
		}
		s := string(b)
		k := domain.Ident
		switch s {
		case "let":
			k = domain.Let
		case "if":
			k = domain.If
		case "else":
			k = domain.Else
		case "return":
			k = domain.Return
		case "true":
			k = domain.True
		case "false":
			k = domain.False
		}
		return domain.Token{Kind: k, Lexeme: s, Pos: start}, nil
	}
	if unicode.IsDigit(r) {
		var b []rune
		dot := false
		for l.offset < len(l.src) && (unicode.IsDigit(l.src[l.offset]) || (!dot && l.src[l.offset] == '.')) {
			if l.src[l.offset] == '.' {
				dot = true
			}
			b = append(b, l.src[l.offset])
			l.advance()
		}
		return domain.Token{Kind: domain.Number, Lexeme: string(b), Pos: start}, nil
	}
	if r == '"' {
		l.advance()
		var b []rune
		for l.offset < len(l.src) && l.src[l.offset] != '"' {
			if l.src[l.offset] == '\\' && l.offset+1 < len(l.src) {
				l.advance()
				b = append(b, l.src[l.offset])
				l.advance()
			} else {
				b = append(b, l.src[l.offset])
				l.advance()
			}
		}
		if l.offset >= len(l.src) {
			return domain.Token{}, fmt.Errorf("unterminated string at %d:%d: %v", start.Line, start.Column, domain.ErrUnterminatedString)
		}
		l.advance()
		return domain.Token{Kind: domain.String, Lexeme: string(b), Pos: start}, nil
	}
	for _, p := range []struct {
		r rune
		k domain.Kind
	}{{'+', domain.Plus}, {'-', domain.Minus}, {'*', domain.Star}, {'/', domain.Slash}, {'(', domain.LParen}, {')', domain.RParen}, {'{', domain.LBrace}, {'}', domain.RBrace}, {',', domain.Comma}, {'.', domain.Dot}, {':', domain.Colon}, {';', domain.Semicolon}} {
		if r == p.r {
			l.advance()
			return domain.Token{Kind: p.k, Lexeme: string(r), Pos: start}, nil
		}
	}
	if l.offset+1 < len(l.src) {
		two := string([]rune{r, l.src[l.offset+1]})
		m := map[string]domain.Kind{"==": domain.EqualEqual, "!=": domain.NotEqual, "<=": domain.LessEqual, ">=": domain.GreaterEqual, "&&": domain.And, "||": domain.Or}
		if k, ok := m[two]; ok {
			l.advance()
			l.advance()
			return domain.Token{Kind: k, Lexeme: two, Pos: start}, nil
		}
	}
	m := map[rune]domain.Kind{'=': domain.Equal, '<': domain.Less, '>': domain.Greater, '!': domain.Not}
	if k, ok := m[r]; ok {
		l.advance()
		return domain.Token{Kind: k, Lexeme: string(r), Pos: start}, nil
	}
	return domain.Token{}, fmt.Errorf("unexpected character %q at %d:%d: %v", r, start.Line, start.Column, domain.ErrUnexpectedRune)
}

func (l *Lexer) All() ([]domain.Token, error) {
	out := make([]domain.Token, 0)
	for len(out) < l.maxTokens {
		t, e := l.Next()
		if e != nil {
			return nil, e
		}
		out = append(out, t)
		if t.Kind == domain.EOF {
			return out, nil
		}
	}
	return nil, &domain.ScanError{Cause: domain.ErrLimit, Pos: l.pos(), Text: "token budget exceeded"}
}
func (l *Lexer) pos() domain.Position {
	return domain.Position{Offset: l.offset, Line: l.line, Column: l.column}
}
func (l *Lexer) tok(k domain.Kind, s string) domain.Token {
	return domain.Token{Kind: k, Lexeme: s, Pos: l.pos()}
}
func (l *Lexer) advance() {
	if l.offset >= len(l.src) {
		return
	}
	if l.src[l.offset] == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.offset++
}
