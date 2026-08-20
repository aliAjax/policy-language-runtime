package application

import (
	"errors"
	"fmt"
	lexadapter "github.com/example/policy-language-runtime/internal/lexer/adapter"
	lexapp "github.com/example/policy-language-runtime/internal/lexer/application"
	"github.com/example/policy-language-runtime/internal/lexer/domain"
	parseradapter "github.com/example/policy-language-runtime/internal/parser/adapter"
	ast "github.com/example/policy-language-runtime/internal/parser/domain"
	"strconv"
)

func ParseSource(src string, limits lexadapter.Limits) (*ast.Program, error) {
	tokens, err := lexapp.NewWithLimits(src, limits).All()
	if err != nil {
		var scan *domain.ScanError
		if errors.As(err, &scan) {
			return nil, parseradapter.NewCause("lexing failed", scan.Pos.Line, scan.Pos.Column, err)
		}
		return nil, parseradapter.NewCause("lexing failed", 1, 1, err)
	}
	return New(tokens, 64).Parse()
}

type Parser struct {
	ts       []domain.Token
	i        int
	maxDepth int
}

func New(ts []domain.Token, maxDepth int) *Parser {
	if maxDepth <= 0 {
		maxDepth = 64
	}
	return &Parser{ts: ts, maxDepth: maxDepth}
}
func (p *Parser) Parse() (*ast.Program, error) {
	out := &ast.Program{}
	for !p.at(domain.EOF) {
		s, e := p.stmt(0)
		if e != nil {
			return nil, e
		}
		out.Statements = append(out.Statements, s)
	}
	return out, nil
}
func (p *Parser) stmt(depth int) (ast.Stmt, error) {
	if depth > p.maxDepth {
		return nil, fmt.Errorf("nesting depth exceeds %d", p.maxDepth)
	}
	if p.at(domain.Let) {
		p.next()
		id := p.next()
		if id.Kind != domain.Ident {
			return nil, p.err("expected identifier", id)
		}
		if !p.at(domain.Equal) {
			return nil, p.err("expected =", p.peek())
		}
		p.next()
		v, e := p.expr(depth + 1)
		if e != nil {
			return nil, e
		}
		p.optional(domain.Semicolon)
		return ast.Let{Name: id.Lexeme, Value: v}, nil
	}
	if p.at(domain.Return) {
		p.next()
		v, e := p.expr(depth + 1)
		if e != nil {
			return nil, e
		}
		p.optional(domain.Semicolon)
		return ast.Return{Value: v}, nil
	}
	if p.at(domain.If) {
		p.next()
		c, e := p.expr(depth + 1)
		if e != nil {
			return nil, e
		}
		b, e := p.block(depth + 1)
		if e != nil {
			return nil, e
		}
		var el *ast.Block
		if p.at(domain.Else) {
			p.next()
			el, e = p.block(depth + 1)
			if e != nil {
				return nil, e
			}
		}
		return ast.If{Cond: c, Then: b, Else: el}, nil
	}
	v, e := p.expr(depth + 1)
	if e != nil {
		return nil, e
	}
	p.optional(domain.Semicolon)
	return ast.ExprStmt{Value: v}, nil
}
func (p *Parser) block(depth int) (*ast.Block, error) {
	if !p.at(domain.LBrace) {
		return nil, p.err("expected {", p.peek())
	}
	p.next()
	b := &ast.Block{}
	for !p.at(domain.RBrace) && !p.at(domain.EOF) {
		s, e := p.stmt(depth + 1)
		if e != nil {
			return nil, e
		}
		b.Statements = append(b.Statements, s)
	}
	if !p.at(domain.RBrace) {
		return nil, p.err("expected }", p.peek())
	}
	p.next()
	return b, nil
}
func (p *Parser) expr(depth int) (ast.Expr, error) { return p.binary(depth, 0) }

var prec = map[domain.Kind]int{domain.Or: 1, domain.And: 2, domain.EqualEqual: 3, domain.NotEqual: 3, domain.Less: 4, domain.LessEqual: 4, domain.Greater: 4, domain.GreaterEqual: 4, domain.Plus: 5, domain.Minus: 5, domain.Star: 6, domain.Slash: 6}

func (p *Parser) binary(depth, min int) (ast.Expr, error) {
	l, e := p.unary(depth)
	if e != nil {
		return nil, e
	}
	for {
		op, ok := prec[p.peek().Kind]
		if !ok || op < min {
			break
		}
		t := p.next()
		r, e := p.binary(depth, op+1)
		if e != nil {
			return nil, e
		}
		l = ast.Binary{Op: t.Lexeme, Left: l, Right: r}
	}
	return l, nil
}
func (p *Parser) unary(depth int) (ast.Expr, error) {
	if p.at(domain.Not) || p.at(domain.Minus) {
		t := p.next()
		v, e := p.unary(depth + 1)
		if e != nil {
			return nil, e
		}
		return ast.Unary{Op: t.Lexeme, Value: v}, nil
	}
	return p.primary(depth)
}
func (p *Parser) primary(depth int) (ast.Expr, error) {
	t := p.next()
	switch t.Kind {
	case domain.Number:
		n, e := strconv.ParseFloat(t.Lexeme, 64)
		if e != nil {
			return nil, e
		}
		return ast.Literal{Value: n}, nil
	case domain.String:
		return ast.Literal{Value: t.Lexeme}, nil
	case domain.True:
		return ast.Literal{Value: true}, nil
	case domain.False:
		return ast.Literal{Value: false}, nil
	case domain.Ident:
		if p.at(domain.LParen) {
			p.next()
			c := ast.Call{Name: t.Lexeme}
			for !p.at(domain.RParen) {
				a, e := p.expr(depth + 1)
				if e != nil {
					return nil, e
				}
				c.Args = append(c.Args, a)
				if !p.optional(domain.Comma) {
					break
				}
			}
			if !p.at(domain.RParen) {
				return nil, p.err("expected )", p.peek())
			}
			p.next()
			return c, nil
		}
		return ast.Variable{Name: t.Lexeme}, nil
	case domain.LParen:
		v, e := p.expr(depth + 1)
		if e != nil {
			return nil, e
		}
		if !p.at(domain.RParen) {
			return nil, p.err("expected )", p.peek())
		}
		p.next()
		return v, nil
	}
	return nil, p.err("unexpected token", t)
}
func (p *Parser) peek() domain.Token {
	if p.i >= len(p.ts) {
		return domain.Token{Kind: domain.EOF}
	}
	return p.ts[p.i]
}
func (p *Parser) next() domain.Token {
	t := p.peek()
	if p.i < len(p.ts) {
		p.i++
	}
	return t
}
func (p *Parser) at(k domain.Kind) bool { return p.peek().Kind == k }
func (p *Parser) optional(k domain.Kind) bool {
	if p.at(k) {
		p.next()
		return true
	}
	return false
}
func (p *Parser) err(s string, t domain.Token) error {
	return parseradapter.NewCause(s, t.Pos.Line, t.Pos.Column, nil)
}
