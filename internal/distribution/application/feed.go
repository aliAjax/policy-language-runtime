package application

import (
	"bytes"
	"context"
	"fmt"
	"github.com/example/policy-language-runtime/internal/distribution/adapter"
	"github.com/example/policy-language-runtime/internal/distribution/domain"
	"github.com/example/policy-language-runtime/internal/platform"
	"sync"
)

type Subscriber func(context.Context) ([]domain.Update, error)

type Feed struct {
	mu     sync.Mutex
	cursor int64
	items  []domain.Update
}

func New() *Feed { return &Feed{} }
func (f *Feed) Publish(u domain.Update) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cursor++
	u.Cursor = f.cursor
	f.items = append(f.items, u)
}
func (f *Feed) Since(ctx context.Context, c int64, limit int) []domain.Update {
	f.mu.Lock()
	defer f.mu.Unlock()
	if limit <= 0 {
		limit = 100
	}
	o := []domain.Update{}
	for _, u := range f.items {
		if u.Cursor > c && len(o) < limit {
			o = append(o, u)
		}
	}
	return o
}

func (f *Feed) Stream(ctx context.Context, subscribers ...Subscriber) (<-chan domain.Update, <-chan error) {
	updates := make(chan domain.Update)
	errorsCh := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	for _, subscriber := range subscribers {
		subscriber := subscriber
		go func() {
			defer wg.Done()
			batch, err := subscriber(ctx)
			if err != nil {
				select {
				case errorsCh <- err:
				case <-ctx.Done():
				}
				return
			}
			for _, update := range batch {
				select {
				case updates <- update.Clone():
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(updates)
		close(errorsCh)
	}()
	return updates, errorsCh
}

func SignedSubscriber(verifier *adapter.Verifier, input <-chan domain.Update) Subscriber {
	return func(ctx context.Context) ([]domain.Update, error) {
		var updates []domain.Update
		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case update, ok := <-input:
				if !ok {
					return updates, nil
				}
				if err := verifier.VerifyUpdate(ctx, update); err != nil {
					return nil, fmt.Errorf("verify update: %w", err)
				}
				updates = append(updates, update.Clone())
			}
		}
	}
}

func PayloadDigest(ctx context.Context, update domain.Update) (string, error) {
	return platform.HashReaderContext(ctx, bytes.NewReader(update.Payload))
}
