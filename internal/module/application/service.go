package application

import (
	"context"
	"fmt"
	"github.com/example/policy-language-runtime/internal/module/domain"
)

type Repository interface {
	Save(context.Context, domain.Module) error
	Versions(context.Context, string) ([]string, error)
}
type Service struct{ Repo Repository }

type BatchRepository interface {
	SaveBatch(context.Context, []domain.Module) error
}

func New(r Repository) *Service { return &Service{Repo: r} }
func (s *Service) Register(ctx context.Context, m domain.Module) error {
	if err := m.Validate(); err != nil {
		return err
	}
	return s.Repo.Save(ctx, m)
}

func (s *Service) RegisterBatch(ctx context.Context, modules []domain.Module) error {
	if len(modules) == 0 {
		return fmt.Errorf("module batch is empty")
	}
	for i, module := range modules {
		if err := s.Register(ctx, module); err != nil {
			return fmt.Errorf("module %d: %w", i, err)
		}
	}
	return nil
}
