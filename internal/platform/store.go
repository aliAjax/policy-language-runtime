package platform

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Package struct {
	Namespace, Name, Version, Channel, Source, Checksum, Status string
	CreatedAt                                                   time.Time
}
type Store interface {
	Put(context.Context, Package) error
	Get(context.Context, string, string, string) (Package, error)
	Latest(context.Context, string, string, string) (Package, error)
	List(context.Context, string, string) ([]Package, error)
}
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]Package
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{data: map[string]Package{}} }
func key(n, m, v string) string    { return n + "/" + m + "/" + v }
func (s *MemoryStore) Put(ctx context.Context, p Package) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key(p.Namespace, p.Name, p.Version)] = p
	return nil
}
func (s *MemoryStore) Get(ctx context.Context, n, m, v string) (Package, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.data[key(n, m, v)]
	if !ok {
		return Package{}, fmt.Errorf("package %s/%s@%s: %v", n, m, v, ErrNotFound)
	}
	return p, nil
}
func (s *MemoryStore) Latest(ctx context.Context, n, m, c string) (Package, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out Package
	for _, p := range s.data {
		if p.Namespace == n && p.Name == m && (c == "" || p.Channel == c) && p.Status == "published" && (out.Version == "" || p.CreatedAt.After(out.CreatedAt)) {
			out = p
		}
	}
	if out.Version == "" {
		return Package{}, fmt.Errorf("published package %s/%s channel %s: %w", n, m, c, ErrNotFound)
	}
	return out, nil
}
func (s *MemoryStore) List(ctx context.Context, n, m string) ([]Package, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := []Package{}
	for _, p := range s.data {
		if p.Namespace == n && p.Name == m {
			o = append(o, p)
		}
	}
	return o, nil
}
