package platform

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	irapp "github.com/example/policy-language-runtime/internal/ir/application"
	ird "github.com/example/policy-language-runtime/internal/ir/domain"
	lex "github.com/example/policy-language-runtime/internal/lexer/application"
	parser "github.com/example/policy-language-runtime/internal/parser/application"
	run "github.com/example/policy-language-runtime/internal/runtime/application"
	"github.com/example/policy-language-runtime/internal/runtime/domain"
	tc "github.com/example/policy-language-runtime/internal/typecheck/application"
	"time"
)

type Service struct {
	Store Store
	Eval  *run.Evaluator
}

func NewService(s Store) *Service { return &Service{Store: s, Eval: run.New()} }
func (s *Service) Compile(ctx context.Context, src string) (any, []string, error) {
	ts, e := lex.New(src, 10000).All()
	if e != nil {
		return nil, nil, fmt.Errorf("lex: %w", e)
	}
	p, e := parser.New(ts, 64).Parse()
	if e != nil {
		return nil, nil, fmt.Errorf("parse: %w", e)
	}
	ds := tc.New().Check(p)
	if len(ds) > 0 {
		return nil, []string{ds[0].Message}, fmt.Errorf("typecheck failed: %s", ds[0].Message)
	}
	prog, e := irapp.New().Compile(p)
	if e != nil {
		return nil, nil, e
	}
	return prog, nil, nil
}
func checksum(src string) string { h := sha256.Sum256([]byte(src)); return hex.EncodeToString(h[:]) }
func (s *Service) Publish(ctx context.Context, n, m, v, ch, src string) (Package, error) {
	p, _, e := s.Compile(ctx, src)
	if e != nil {
		return Package{}, e
	}
	_ = p
	now := timeNow()
	pkg := Package{Namespace: n, Name: m, Version: v, Channel: ch, Source: src, Checksum: checksum(src), Status: "published", CreatedAt: now}
	if e := s.Store.Put(ctx, pkg); e != nil {
		return Package{}, e
	}
	return pkg, nil
}
func (s *Service) Decide(ctx context.Context, n, m, v, ch string, input map[string]any) (domain.Result, error) {
	var p Package
	var e error
	if v != "" {
		p, e = s.Store.Get(ctx, n, m, v)
	} else {
		p, e = s.Store.Latest(ctx, n, m, ch)
	}
	if e != nil {
		return domain.Result{}, fmt.Errorf("resolve policy %s/%s: %v", n, m, e)
	}
	raw, _, e := s.Compile(ctx, p.Source)
	if e != nil {
		return domain.Result{}, fmt.Errorf("compile published policy %s/%s: %w", n, m, e)
	}
	return s.Eval.Eval(ctx, raw.(irProgram), input)
}

type irProgram = ird.Program

var timeNow = func() time.Time { return time.Now() }
