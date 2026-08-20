package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/policy-language-runtime/internal/distribution/adapter"
	app "github.com/example/policy-language-runtime/internal/distribution/application"
	"github.com/example/policy-language-runtime/internal/distribution/domain"
)

func collect(updates <-chan domain.Update) []domain.Update {
	var out []domain.Update
	for update := range updates {
		out = append(out, update)
	}
	return out
}

func TestFeedWaitsForAllSubscribers(t *testing.T) {
	start := make(chan struct{})
	first := func(context.Context) ([]domain.Update, error) { <-start; return []domain.Update{{Version: "1"}}, nil }
	second := func(context.Context) ([]domain.Update, error) { <-start; return []domain.Update{{Version: "2"}}, nil }
	updates, errs := app.New().Stream(context.Background(), first, second)
	close(start)
	got := collect(updates)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(got) != 2 {
		t.Fatalf("want both subscribers, got %#v", got)
	}
}

func TestFeedClosesAfterVerifierError(t *testing.T) {
	input := make(chan domain.Update, 1)
	input <- domain.Update{Payload: []byte("payload"), Signature: "bad"}
	close(input)
	updates, errs := app.New().Stream(context.Background(), app.SignedSubscriber(adapter.NewVerifier([]byte("secret")), input))
	if got := collect(updates); len(got) != 0 {
		t.Fatalf("bad update was published: %#v", got)
	}
	if err := <-errs; err == nil {
		t.Fatal("signature failure was lost")
	}
}

func TestFeedErrorChannelCannotDeadlock(t *testing.T) {
	want := errors.New("subscriber failed")
	updates, errs := app.New().Stream(context.Background(), func(context.Context) ([]domain.Update, error) { return nil, want })
	done := make(chan struct{})
	go func() { collect(updates); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("result channel waited for error consumer")
	}
	if err := <-errs; !errors.Is(err, want) {
		t.Fatalf("want subscriber error, got %v", err)
	}
}

func TestFeedPublishesCompleteSnapshot(t *testing.T) {
	secret := []byte("secret")
	verifier := adapter.NewVerifier(secret)
	payload := []byte("policy-data")
	input := make(chan domain.Update, 1)
	input <- domain.Update{Version: "1.0.0", Payload: payload, Signature: verifier.Sign(payload)}
	close(input)
	updates, _ := app.New().Stream(context.Background(), app.SignedSubscriber(verifier, input))
	got := collect(updates)
	payload[0] = 'X'
	secret[0] = 'X'
	if len(got) != 1 || string(got[0].Payload) != "policy-data" || !verifier.Verify(got[0].Payload, got[0].Signature) {
		t.Fatalf("published update is not detached: %#v", got)
	}
}

func TestFeedChecksumHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := app.PayloadDigest(ctx, domain.Update{Payload: make([]byte, 1024)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context cancellation, got %v", err)
	}
}
