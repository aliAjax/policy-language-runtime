package application_test

import (
	"context"
	"testing"

	app "github.com/example/policy-language-runtime/internal/namespace/application"
	"github.com/example/policy-language-runtime/internal/namespace/domain"
	"github.com/example/policy-language-runtime/internal/namespace/infrastructure"
	"github.com/example/policy-language-runtime/internal/platform"
)

func namespaceService(t *testing.T, state domain.State) (*app.Service, *infrastructure.Memory) {
	t.Helper()
	repo := infrastructure.New()
	if err := repo.Save(context.Background(), domain.Namespace{ID: "tenant-a", Name: "Tenant A", State: state, Active: state == domain.Active}); err != nil {
		t.Fatal(err)
	}
	return app.New(repo), repo
}

func TestNamespaceRecoveryReachesActive(t *testing.T) {
	svc, repo := namespaceService(t, domain.Suspended)
	if err := svc.Recover(context.Background(), "tenant-a"); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.Find(context.Background(), "tenant-a")
	if got.CurrentState() != domain.Active || !got.Active {
		t.Fatalf("recovery stopped in %s", got.CurrentState())
	}
}

func TestNamespaceRejectsIllegalTransition(t *testing.T) {
	svc, _ := namespaceService(t, domain.Active)
	if err := svc.ChangeState(context.Background(), "tenant-a", domain.Draft); err == nil {
		t.Fatal("active -> draft must be rejected")
	}
}

func TestNamespaceStateReadIsConsistent(t *testing.T) {
	_, repo := namespaceService(t, domain.Active)
	ok, err := repo.CompareAndSwap(context.Background(), "tenant-a", domain.Active, domain.Suspended)
	if err != nil || !ok {
		t.Fatalf("first transition failed: ok=%v err=%v", ok, err)
	}
	ok, err = repo.CompareAndSwap(context.Background(), "tenant-a", domain.Active, domain.Archived)
	if err != nil || ok {
		t.Fatalf("stale transition accepted: ok=%v err=%v", ok, err)
	}
	got, _ := repo.Find(context.Background(), "tenant-a")
	if got.CurrentState() != domain.Suspended {
		t.Fatalf("want suspended, got %s", got.CurrentState())
	}
}

func TestRecoveredReleaseCanEvaluate(t *testing.T) {
	r := platform.Release{Version: "1.0.0", Channel: "stable", State: "suspended"}
	r.Recover()
	if !r.CanEvaluate() || r.State != "active" {
		t.Fatalf("recovered release is not active: %#v", r)
	}
}
