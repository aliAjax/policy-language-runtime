package application

import (
	"context"
	ir "github.com/example/policy-language-runtime/internal/ir/domain"
	"github.com/example/policy-language-runtime/internal/runtime/application"
	"github.com/example/policy-language-runtime/internal/simulation/domain"
)

type Service struct {
	Eval *application.Evaluator
}

func New(e *application.Evaluator) *Service { return &Service{Eval: e} }
func (s *Service) Compare(ctx context.Context, a, b irProgram, samples []domain.Sample) domain.Report {
	r := domain.Report{Total: len(samples), Samples: make([]domain.Sample, len(samples))}
	for i := range samples {
		r.Samples[i] = samples[i].Clone()
	}
	for _, x := range samples {
		select {
		case <-ctx.Done():
			r.Errors++
			return r.Clone()
		default:
		}
		snapshot := x.Clone()
		ra, e1 := s.Eval.Eval(ctx, a.Clone(), snapshot.Input)
		rb, e2 := s.Eval.Eval(ctx, b.Clone(), snapshot.Input)
		if e1 != nil || e2 != nil {
			r.Errors++
			continue
		}
		if ra.Decision != rb.Decision {
			r.Changes = append(r.Changes, domain.Change{SampleID: x.ID, Before: string(ra.Decision), After: string(rb.Decision), Kind: "decision"})
		}
	}
	return r
}

type irProgram = ir.Program
