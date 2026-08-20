package storage

import (
	"context"
	"fmt"
	"io"
)

type MigrationStep struct {
	Name  string
	Apply func(context.Context) (io.Closer, error)
}

type Migrator struct {
	Steps      []string
	Operations []MigrationStep
}

func DefaultMigrations() *Migrator {
	return &Migrator{Steps: []string{"create namespaces", "create modules", "create policy_versions", "create release_channels", "create simulation_samples"}}
}
func (m *Migrator) Run(ctx context.Context) error {
	if len(m.Steps) == 0 && len(m.Operations) == 0 {
		return fmt.Errorf("no migrations")
	}
	for _, step := range m.Operations {
		if err := runMigrationStep(ctx, step); err != nil {
			return err
		}
	}
	return nil
}

func runMigrationStep(ctx context.Context, step MigrationStep) error {
	if step.Apply == nil {
		return fmt.Errorf("migration %s has no operation", step.Name)
	}
	resource, err := step.Apply(ctx)
	if err != nil {
		return fmt.Errorf("migration %s: %w", step.Name, err)
	}
	if resource == nil {
		return nil
	}
	if err := resource.Close(); err != nil {
		return fmt.Errorf("close migration %s resource: %w", step.Name, err)
	}
	return nil
}
