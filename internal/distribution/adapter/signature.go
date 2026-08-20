package adapter

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/example/policy-language-runtime/internal/distribution/domain"
)

type Verifier struct {
	mu     sync.RWMutex
	secret []byte
}

func NewVerifier(secret []byte) *Verifier {
	return &Verifier{secret: secret}
}

func (v *Verifier) Sign(data []byte) string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	h := sha256.New()
	h.Write(v.secret)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
func (v *Verifier) Verify(data []byte, signature string) bool {
	want, err := hex.DecodeString(v.Sign(data))
	if err != nil {
		return false
	}
	got, err := hex.DecodeString(signature)
	return err == nil && len(want) == len(got) && subtle.ConstantTimeCompare(want, got) == 1
}

func (v *Verifier) VerifyUpdate(ctx context.Context, update domain.Update) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if v == nil || !v.Verify(update.Payload, update.Signature) {
		return fmt.Errorf("update signature rejected")
	}
	return nil
}
