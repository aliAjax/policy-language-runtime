package application

import (
	"fmt"
	"github.com/example/policy-language-runtime/internal/ir/domain"
	"github.com/example/policy-language-runtime/internal/optimizer/adapter"
)

type Optimizer struct {
	Folds int
	deps  []string
}

func New() *Optimizer { return &Optimizer{} }
func (o *Optimizer) Optimize(p domain.Program) domain.Program {
	o.Folds = 0
	o.deps = append([]string(nil), p.Deps...)
	out := p
	code := out.Code[:0]
	for _, instruction := range out.Code {
		if instruction.Op == domain.Binary && len(code) >= 2 && code[len(code)-1].Op == domain.LoadConst && code[len(code)-2].Op == domain.LoadConst {
			if value, ok := fold(instruction.Arg, code[len(code)-2].Value, code[len(code)-1].Value); ok {
				code = code[:len(code)-2]
				code = append(code, domain.Instr{Op: domain.LoadConst, Value: value})
				o.Folds++
				continue
			}
		}
		code = append(code, instruction)
	}
	out.Code = code
	return out
}
func (o *Optimizer) Stats() string { return fmt.Sprintf("constant_folds=%d", o.Folds) }
func (o *Optimizer) Report() adapter.Report {
	return adapter.Report{ConstantsFolded: o.Folds, Dependencies: len(o.deps), DependencyNames: append([]string(nil), o.deps...)}
}

func fold(op string, left, right any) (any, bool) {
	a, aok := left.(float64)
	b, bok := right.(float64)
	if !aok || !bok {
		return nil, false
	}
	switch op {
	case "+":
		return a + b, true
	case "-":
		return a - b, true
	case "*":
		return a * b, true
	case "/":
		if b != 0 {
			return a / b, true
		}
	}
	return nil, false
}
