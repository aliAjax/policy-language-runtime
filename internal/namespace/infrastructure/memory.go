package infrastructure

import (
	"context"
	"fmt"
	"github.com/example/policy-language-runtime/internal/namespace/domain"
	"sync"
)

type Memory struct {
	mu    sync.RWMutex
	items map[string]domain.Namespace
}

func New() *Memory { return &Memory{items: map[string]domain.Namespace{}} }
func (m *Memory) Save(ctx context.Context, n domain.Namespace) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[n.ID] = n
	return nil
}
func (m *Memory) Find(ctx context.Context, id string) (domain.Namespace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.items[id]
	if !ok {
		return domain.Namespace{}, fmt.Errorf("namespace %s not found", id)
	}
	return n, nil
}

func (m *Memory) CompareAndSwap(ctx context.Context, id string, from, to domain.State) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.items[id]
	if !ok {
		return false, fmt.Errorf("namespace %s not found", id)
	}
	if err := n.Transition(to); err != nil {
		return false, err
	}
	m.items[id] = n
	return true, nil
}
