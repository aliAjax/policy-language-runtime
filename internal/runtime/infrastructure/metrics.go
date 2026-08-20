package infrastructure

import "sync"

type Metrics struct {
	mu               sync.RWMutex
	Requests, Errors uint64
}

func (m *Metrics) IncRequest() { m.Record(false) }
func (m *Metrics) IncError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors++
}
func (m *Metrics) Record(failed bool) {
	m.Requests++
	if failed {
		m.Errors++
	}
}
func (m *Metrics) Snapshot() (uint64, uint64) {
	return m.Requests, m.Errors
}
