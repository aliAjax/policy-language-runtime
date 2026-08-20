package adapter

import "fmt"

type Diagnostic struct {
	Message              string
	Offset, Line, Column int
	Hint                 string
	Cause                error
}

func New(msg string, line, col int) Diagnostic {
	return Diagnostic{Message: msg, Line: line, Column: col}
}

func NewCause(msg string, line, col int, cause error) *Diagnostic {
	return &Diagnostic{Message: msg, Line: line, Column: col, Cause: cause}
}

func (d *Diagnostic) Error() string {
	if d == nil {
		return "parser diagnostic"
	}
	return fmt.Sprintf("%s at %d:%d", d.Message, d.Line, d.Column)
}

func (d *Diagnostic) Unwrap() error {
	return nil
}
