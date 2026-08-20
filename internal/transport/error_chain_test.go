package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/policy-language-runtime/internal/platform"
	"github.com/example/policy-language-runtime/internal/transport"
)

func TestMissingPolicyPreservesErrorChain(t *testing.T) {
	_, err := platform.NewMemoryStore().Get(context.Background(), "acme", "access", "9.9.9")
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("missing package must preserve ErrNotFound, got %v", err)
	}
}

func TestMissingPolicyServiceClassification(t *testing.T) {
	svc := platform.NewService(platform.NewMemoryStore())
	_, err := svc.Decide(context.Background(), "acme", "access", "9.9.9", "", map[string]any{})
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("service must preserve missing classification, got %v", err)
	}
}

func TestMissingPolicyHTTPStatus(t *testing.T) {
	svc := platform.NewService(platform.NewMemoryStore())
	h := transport.New(svc, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler()
	body := bytes.NewBufferString(`{"namespace":"acme","module":"access","version":"9.9.9","input":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/decide", body)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d: %s", rr.Code, rr.Body.String())
	}
	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil || got["code"] != "not_found" {
		t.Fatalf("want not_found body, got %q (%v)", rr.Body.String(), err)
	}
}

func TestMissingPolicyDoesNotRetry(t *testing.T) {
	err := errors.Join(errors.New("lookup failed"), platform.ErrNotFound)
	if platform.Retryable(err) {
		t.Fatal("missing policy must not be retried")
	}
}
