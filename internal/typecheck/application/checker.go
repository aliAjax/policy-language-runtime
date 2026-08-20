package application

import (
	"fmt"
	builtin "github.com/example/policy-language-runtime/internal/builtin/domain"
	"github.com/example/policy-language-runtime/internal/parser/domain"
	t "github.com/example/policy-language-runtime/internal/typecheck/domain"
)

type Checker struct {
	vars  map[string]t.Type
	diags []t.Diagnostic
	funcs builtin.Registry
}

func New() *Checker { return &Checker{vars: map[string]t.Type{}} }
func NewWithRegistry(funcs builtin.Registry) *Checker {
	return &Checker{vars: map[string]t.Type{}, funcs: funcs}
}
func (c *Checker) Check(p *domain.Program) []t.Diagnostic {
	c.diags = nil
	if c.funcs != nil && !c.funcs.Ready() {
		c.diags = append(c.diags, t.Diagnostic{Code: "registry_unavailable", Message: builtin.ErrRegistryUnavailable.Error(), Severity: "error"})
		return append([]t.Diagnostic(nil), c.diags...)
	}
	if p == nil {
		c.diags = append(c.diags, t.Diagnostic{Code: "nil_program", Message: "program is nil", Severity: "error"})
		return append([]t.Diagnostic(nil), c.diags...)
	}
	for _, s := range p.Statements {
		c.stmt(s)
	}
	return c.diags
}
func (c *Checker) stmt(s domain.Stmt) {
	switch x := s.(type) {
	case domain.Let:
		v := c.expr(x.Value)
		c.vars[x.Name] = v
	case domain.Return:
		c.expr(x.Value)
	case domain.ExprStmt:
		c.expr(x.Value)
	case domain.If:
		if c.expr(x.Cond) != t.Bool {
			c.diags = append(c.diags, t.Diagnostic{Message: fmt.Sprintf("condition must be bool, got %s", c.expr(x.Cond)), Severity: "error"})
		}
		for _, z := range x.Then.Statements {
			c.stmt(z)
		}
		if x.Else != nil {
			for _, z := range x.Else.Statements {
				c.stmt(z)
			}
		}
	}
}
func (c *Checker) expr(e domain.Expr) t.Type {
	switch x := e.(type) {
	case domain.Literal:
		switch x.Value.(type) {
		case bool:
			return t.Bool
		case float64:
			return t.Number
		case string:
			return t.String
		}
		return t.Any
	case domain.Variable:
		if v, ok := c.vars[x.Name]; ok {
			return v
		}
		// Runtime input variables are dynamically bound; retain Any and defer missing-field handling to evaluation.
		return t.Any
	case domain.Unary:
		v := c.expr(x.Value)
		if x.Op == "!" && v == t.Bool {
			return t.Bool
		}
		if x.Op == "-" && v == t.Number {
			return t.Number
		}
		c.diags = append(c.diags, t.Diagnostic{Message: "invalid unary operand", Severity: "error"})
		return t.Any
	case domain.Binary:
		l, r := c.expr(x.Left), c.expr(x.Right)
		if x.Op == "&&" || x.Op == "||" {
			if l == t.Bool && r == t.Bool {
				return t.Bool
			}
		}
		if x.Op == "+" || x.Op == "-" || x.Op == "*" || x.Op == "/" {
			if l == t.Number && r == t.Number {
				return t.Number
			}
		}
		if x.Op == "==" || x.Op == "!=" || x.Op == "<" || x.Op == "<=" || x.Op == ">" || x.Op == ">=" {
			return t.Bool
		}
		c.diags = append(c.diags, t.Diagnostic{Message: "incompatible operands", Severity: "error"})
		return t.Any
	case domain.Call:
		for _, a := range x.Args {
			c.expr(a)
		}
		if x.Name == "contains" || x.Name == "matches" {
			return t.Bool
		}
		if x.Name == "now" {
			return t.Number
		}
		if c.funcs != nil {
			return t.Any
		}
		return t.Any
	}
	return t.Any
}
