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
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	for i, module := range modules {
		if err := module.Validate(); err != nil {
			return fmt.Errorf("module %d: %w", i, err)
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	nextItems := make(map[string]domain.Module, len(m.items)+len(modules))
	for key, module := range m.items {
		nextItems[key] = module.Clone()
	}
	nextVersions := make(map[string][]string, len(m.versions)+len(modules))
	for key, versions := range m.versions {
		nextVersions[key] = append([]string(nil), versions...)
	}
	for _, module := range modules {
		nextItems[module.Key()] = module.Clone()
		if module.Version != "" {
			nextVersions[module.Key()] = append(nextVersions[module.Key()], module.Version)
		}
	}
	m.items, m.versions = nextItems, nextVersions
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
