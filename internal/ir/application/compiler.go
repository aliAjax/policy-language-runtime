package application

import (
	"fmt"
	"github.com/example/policy-language-runtime/internal/ir/domain"
	ast "github.com/example/policy-language-runtime/internal/parser/domain"
)

type Compiler struct{ p domain.Program }

func New() *Compiler { return &Compiler{} }
func (c *Compiler) Compile(src *ast.Program) (domain.Program, error) {
	c.p = domain.Program{}
	for _, s := range src.Statements {
		if e := c.stmt(s); e != nil {
			return domain.Program{}, e
		}
	}
	return c.p, nil
}
func (c *Compiler) stmt(s ast.Stmt) error {
	switch x := s.(type) {
	case ast.Let:
		if e := c.expr(x.Value); e != nil {
			return e
		}
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.LoadVar, Arg: x.Name})
	case ast.Return:
		if e := c.expr(x.Value); e != nil {
			return e
		}
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.Return})
	case ast.ExprStmt:
		if e := c.expr(x.Value); e != nil {
			return e
		}
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.Return})
	case ast.If:
		if e := c.expr(x.Cond); e != nil {
			return e
		}
		j := len(c.p.Code)
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.JumpIfFalse})
		for _, z := range x.Then.Statements {
			if e := c.stmt(z); e != nil {
				return e
			}
		}
		end := len(c.p.Code)
		c.p.Code[j].Target = end
		for _, z := range func() []ast.Stmt {
			if x.Else != nil {
				return x.Else.Statements
			}
			return nil
		}() {
			if e := c.stmt(z); e != nil {
				return e
			}
		}
	}
	return nil
}
func (c *Compiler) expr(e ast.Expr) error {
	switch x := e.(type) {
	case ast.Literal:
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.LoadConst, Value: x.Value})
	case ast.Variable:
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.LoadVar, Arg: x.Name})
		c.p.Deps = append(c.p.Deps, x.Name)
	case ast.Unary:
		if e := c.expr(x.Value); e != nil {
			return e
		}
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.Unary, Arg: x.Op})
	case ast.Binary:
		if e := c.expr(x.Left); e != nil {
			return e
		}
		if e := c.expr(x.Right); e != nil {
			return e
		}
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.Binary, Arg: x.Op})
	case ast.Call:
		for _, a := range x.Args {
			if e := c.expr(a); e != nil {
				return e
			}
		}
		c.p.Code = append(c.p.Code, domain.Instr{Op: domain.Call, Arg: fmt.Sprintf("%s:%d", x.Name, len(x.Args))})
	}
	return nil
}
