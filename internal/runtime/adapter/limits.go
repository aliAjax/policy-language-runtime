package adapter

import (
	"context"
	"time"
)

type Budget struct {
	Deadline      time.Time
	Steps, Memory int
	Trace         []string
}

func (b *Budget) Record(step string) {
	if b == nil || step == "" {
		return
	}
	b.Trace = append(b.Trace, step)
}

func (b Budget) Snapshot() Budget {
	b.Trace = append([]string(nil), b.Trace...)
	return b
}

func NewBudget(d time.Duration, steps int) Budget {
	return Budget{Deadline: time.Now().Add(d), Steps: steps, Memory: 1 << 20}
}
func (b Budget) Check(ctx context.Context, n int) error {
	if n > b.Steps {
		return context.DeadlineExceeded
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
