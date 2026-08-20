package infrastructure

import (
	"context"
	"fmt"
	"github.com/example/policy-language-runtime/internal/module/domain"
	"sync"
)

type Memory struct {
	mu       sync.RWMutex
	items    map[string]domain.Module
	versions map[string][]string
}

func New() *Memory {
	return &Memory{items: map[string]domain.Module{}, versions: map[string][]string{}}
}
func (m *Memory) Save(ctx context.Context, x domain.Module) error {
	if err := x.Validate(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[x.Key()] = x.Clone()
	if x.Version != "" {
		m.versions[x.Key()] = append(m.versions[x.Key()], x.Version)
	}
	return nil
}

func (m *Memory) SaveBatch(ctx context.Context, modules []domain.Module) error {
	for i, module := range modules {
		if err := m.Save(ctx, module); err != nil {
			return fmt.Errorf("module %d: %w", i, err)
		}
	}
	return nil
}
func (m *Memory) Versions(ctx context.Context, k string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.versions[k]...), nil
}

func (m *Memory) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.items[key]
	return ok
}
