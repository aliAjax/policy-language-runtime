package application

import (
	"fmt"
	explain "github.com/example/policy-language-runtime/internal/explain/domain"
	runtime "github.com/example/policy-language-runtime/internal/runtime/domain"
)

type Service struct {
	MaxDepth int
	MaxBytes int
}

func New() *Service { return &Service{MaxDepth: 8, MaxBytes: 4096} }
func (s *Service) Build(r runtime.Result) explain.Explanation {
	v := r.Value
	if s.MaxDepth < 1 {
		return explain.Explanation{Decision: string(r.Decision), Reason: r.Reason, Version: r.Version, Truncated: true}
	}
	children := make([]explain.Node, 0, len(r.Steps))
	for i, step := range r.Steps {
		children = append(children, explain.Node{Path: fmt.Sprintf("steps[%d]", i), Value: step})
	}
	out := explain.Explanation{Decision: string(r.Decision), Reason: r.Reason, Version: r.Version, Root: explain.Node{Path: "result", Value: v, Children: children}, Truncated: false}
	return out
}
func (s *Service) Summary(r runtime.Result) string {
	return fmt.Sprintf("decision=%s reason=%s", r.Decision, r.Reason)
}
