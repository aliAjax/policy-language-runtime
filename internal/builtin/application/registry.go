package application

import (
	"fmt"
	"github.com/example/policy-language-runtime/internal/builtin/domain"
	"sync"
)

type Registry struct {
	mu    sync.RWMutex
	items map[string]domain.Signature
}

func New() *Registry            { return &Registry{items: map[string]domain.Signature{}} }
func (r *Registry) Ready() bool { return r.items != nil }
func (r *Registry) Register(s domain.Signature) error {
	if r == nil {
		return domain.ErrRegistryUnavailable
	}
	if s.Name == "" {
		return fmt.Errorf("empty function name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.items == nil {
		r.items = map[string]domain.Signature{}
	}
	if _, ok := r.items[s.Name]; ok {
		return fmt.Errorf("function exists")
	}
	r.items[s.Name] = s
	return nil
}
func (r *Registry) Lookup(n string) (domain.Signature, bool) {
	if r == nil {
		return domain.Signature{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.items[n]
	return s, ok
}
