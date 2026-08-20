package application

import (
	"context"
	"fmt"
	"github.com/example/policy-language-runtime/internal/namespace/domain"
)

type Repository interface {
	Save(context.Context, domain.Namespace) error
	Find(context.Context, string) (domain.Namespace, error)
	CompareAndSwap(context.Context, string, domain.State, domain.State) (bool, error)
}
type Service struct{ Repo Repository }

func New(r Repository) *Service { return &Service{Repo: r} }
func (s *Service) Create(ctx context.Context, n domain.Namespace) error {
	if !n.Valid() {
		return fmt.Errorf("invalid namespace")
	}
	return s.Repo.Save(ctx, n)
}

func (s *Service) ChangeState(ctx context.Context, id string, next domain.State) error {
	current, err := s.Repo.Find(ctx, id)
	if err != nil {
		return err
	}
	from := current.CurrentState()
	if !domain.CanTransition(from, next) {
		return fmt.Errorf("change namespace %s state: %w", id, fmt.Errorf("transition %s -> %s rejected", from, next))
	}
	swapped, err := s.Repo.CompareAndSwap(ctx, id, from, next)
	if err != nil {
		return err
	}
	if !swapped {
		return fmt.Errorf("namespace %s changed concurrently", id)
	}
	return nil
}

func (s *Service) Recover(ctx context.Context, id string) error {
	// Recovery drives a suspended namespace through the transient
	// Recovering state and on to Active in two atomic steps. Doing both
	// in one call keeps callers from observing a namespace stuck in
	// Recovering, which is the inconsistency the API exposed.
	for _, next := range []domain.State{domain.Recovering, domain.Active} {
		current, err := s.Repo.Find(ctx, id)
		if err != nil {
			return err
		}
		from := current.CurrentState()
		if !domain.CanTransition(from, next) {
			return fmt.Errorf("change namespace %s state: %w", id, fmt.Errorf("transition %s -> %s rejected", from, next))
		}
		swapped, err := s.Repo.CompareAndSwap(ctx, id, from, next)
		if err != nil {
			return err
		}
		if !swapped {
			return fmt.Errorf("namespace %s changed concurrently", id)
		}
	}
	return nil
}
