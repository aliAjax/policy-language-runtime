package infrastructure

import (
	"github.com/example/policy-language-runtime/internal/ir/domain"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]domain.Program
}

func New() *Cache { return &Cache{items: map[string]domain.Program{}} }
func (c *Cache) Get(k string) (domain.Program, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.items[k]
	if !ok {
		return domain.Program{}, false
	}
	return p.Clone(), ok
}
func (c *Cache) Put(k string, p domain.Program) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = p.Clone()
}
