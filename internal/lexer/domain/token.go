package domain

import (
	"errors"
	"fmt"
)

var (
	ErrLimit              = errors.New("lexer limit exceeded")
	ErrUnexpectedRune     = errors.New("unexpected rune")
	ErrUnterminatedString = errors.New("unterminated string")
)

type Kind string

const (
	EOF          Kind = "eof"
	Ident        Kind = "ident"
	Number       Kind = "number"
	String       Kind = "string"
	True         Kind = "true"
	False        Kind = "false"
	Let          Kind = "let"
	If           Kind = "if"
	Else         Kind = "else"
	Return       Kind = "return"
	Plus         Kind = "+"
	Minus        Kind = "-"
	Star         Kind = "*"
	Slash        Kind = "/"
	Equal        Kind = "="
	EqualEqual   Kind = "=="
	NotEqual     Kind = "!="
	Less         Kind = "<"
	LessEqual    Kind = "<="
	Greater      Kind = ">"
	GreaterEqual Kind = ">="
	And          Kind = "&&"
	Or           Kind = "||"
	Not          Kind = "!"
	LParen       Kind = "("
	RParen       Kind = ")"
	LBrace       Kind = "{"
	RBrace       Kind = "}"
	Comma        Kind = ","
	Dot          Kind = "."
	Colon        Kind = ":"
	Semicolon    Kind = ";"
)

type Position struct{ Offset, Line, Column int }

type ScanError struct {
	Cause error
	Pos   Position
	Text  string
}

func (e *ScanError) Error() string {
	if e == nil {
		return "lexer error"
	}
	return fmt.Sprintf("%s at %d:%d", e.Text, e.Pos.Line, e.Pos.Column)
}

func (e *ScanError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type Token struct {
	Kind   Kind
	Lexeme string
	Pos    Position
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%d:%d", t.Kind, t.Lexeme, t.Pos.Line, t.Pos.Column)
}
