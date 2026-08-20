package application

import (
	"context"
	"fmt"
	"github.com/example/policy-language-runtime/internal/ir/domain"
	r "github.com/example/policy-language-runtime/internal/runtime/domain"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Function func(context.Context, []any) (any, error)
type Evaluator struct {
	mu       sync.RWMutex
	funcs    map[string]Function
	MaxSteps int
	MaxStack int
}

func New() *Evaluator {
	e := &Evaluator{funcs: map[string]Function{}, MaxSteps: 10000, MaxStack: 256}
	e.Register("contains", func(_ context.Context, a []any) (any, error) {
		if len(a) != 2 {
			return false, nil
		}
		return strings.Contains(fmt.Sprint(a[0]), fmt.Sprint(a[1])), nil
	})
	e.Register("matches", func(_ context.Context, a []any) (any, error) {
		if len(a) != 2 {
			return false, nil
		}
		return strings.Contains(fmt.Sprint(a[0]), fmt.Sprint(a[1])), nil
	})
	e.Register("now", func(_ context.Context, a []any) (any, error) { return float64(time.Now().Unix()), nil })
	return e
}
func (e *Evaluator) Register(n string, f Function) {
	e.funcs[n] = f
}

func (e *Evaluator) function(name string) (Function, bool) {
	f, ok := e.funcs[name]
	return f, ok
}
func (e *Evaluator) Eval(ctx context.Context, p domain.Program, input map[string]any) (r.Result, error) {
	stack := []any{}
	vars := input
	steps := 0
	for pc := 0; pc < len(p.Code); pc++ {
		steps++
		if steps > e.MaxSteps {
			return r.Result{Decision: r.Indeterminate, Reason: "step budget exceeded"}, fmt.Errorf("step budget exceeded")
		}
		select {
		case <-ctx.Done():
			return r.Result{Decision: r.Indeterminate, Reason: "cancelled"}, ctx.Err()
		default:
		}
		in := p.Code[pc]
		switch in.Op {
		case domain.LoadConst:
			stack = append(stack, in.Value)
		case domain.LoadVar:
			v, ok := vars[in.Arg]
			if !ok {
				return r.Result{Decision: r.Indeterminate, Reason: "missing variable " + in.Arg}, nil
			}
			stack = append(stack, v)
		case domain.Unary:
			if len(stack) < 1 {
				return r.Result{}, fmt.Errorf("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if in.Arg == "!" {
				stack = append(stack, !truth(v))
			} else if n, ok := num(v); ok {
				stack = append(stack, -n)
			} else {
				return r.Result{Decision: r.Indeterminate, Reason: "invalid unary"}, nil
			}
		case domain.Binary:
			if len(stack) < 2 {
				return r.Result{}, fmt.Errorf("stack underflow")
			}
			b, a := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			v, err := binary(in.Arg, a, b)
			if err != nil {
				return r.Result{Decision: r.Indeterminate, Reason: err.Error()}, nil
			}
			stack = append(stack, v)
		case domain.Call:
			name := in.Arg
			parts := strings.Split(name, ":")
			argc, _ := strconv.Atoi(parts[1])
			if len(stack) < argc {
				return r.Result{}, fmt.Errorf("stack underflow")
			}
			args := append([]any(nil), stack[len(stack)-argc:]...)
			stack = stack[:len(stack)-argc]
			f, ok := e.function(parts[0])
			if !ok {
				return r.Result{Decision: r.Indeterminate, Reason: "unknown function"}, nil
			}
			v, err := f(ctx, args)
			if err != nil {
				return r.Result{Decision: r.Indeterminate, Reason: err.Error()}, nil
			}
			stack = append(stack, v)
		case domain.JumpIfFalse:
			if len(stack) < 1 {
				return r.Result{}, fmt.Errorf("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !truth(v) {
				pc = in.Target - 1
			}
		case domain.Return:
			if len(stack) == 0 {
				return r.Result{Decision: r.Indeterminate, Reason: "empty result"}, nil
			}
			v := stack[len(stack)-1]
			if truth(v) {
				return r.Result{Decision: r.Allow, Value: v, Steps: []string{fmt.Sprint(v)}}, nil
			}
			return r.Result{Decision: r.Deny, Value: v, Reason: "policy returned false", Steps: []string{fmt.Sprint(v)}}, nil
		}
	}
	return r.Result{Decision: r.Indeterminate, Reason: "no return"}, nil
}
func truth(v any) bool { b, ok := v.(bool); return ok && b }
func num(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}
func binary(op string, a, b any) (any, error) {
	if op == "&&" {
		return truth(a) && truth(b), nil
	}
	if op == "||" {
		return truth(a) || truth(b), nil
	}
	if op == "==" {
		return fmt.Sprint(a) == fmt.Sprint(b), nil
	}
	if op == "!=" {
		return fmt.Sprint(a) != fmt.Sprint(b), nil
	}
	x, xok := num(a)
	y, yok := num(b)
	if op == "+" && xok && yok {
		return x + y, nil
	}
	if op == "-" && xok && yok {
		return x - y, nil
	}
	if op == "*" && xok && yok {
		return x * y, nil
	}
	if op == "/" && xok && yok && y != 0 {
		return x / y, nil
	}
	if xok && yok {
		switch op {
		case "<":
			return x < y, nil
		case "<=":
			return x <= y, nil
		case ">":
			return x > y, nil
		case ">=":
			return x >= y, nil
		}
	}
	return nil, fmt.Errorf("unsupported operation %s", op)
}
