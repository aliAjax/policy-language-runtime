package application_test

import (
	"context"
	"errors"
	"io"
	"testing"

	moduleapp "github.com/example/policy-language-runtime/internal/module/application"
	"github.com/example/policy-language-runtime/internal/module/domain"
	"github.com/example/policy-language-runtime/internal/module/infrastructure"
	"github.com/example/policy-language-runtime/internal/storage"
)

type fakeTx struct {
	commits, rollbacks     int
	commitErr, rollbackErr error
}

func (f *fakeTx) Commit() error   { f.commits++; return f.commitErr }
func (f *fakeTx) Rollback() error { f.rollbacks++; return f.rollbackErr }

func TestBatchFailureRollsBack(t *testing.T) {
	tx := &fakeTx{}
	want := errors.New("validation failed")
	err := storage.WithTransaction(context.Background(), func(context.Context) (storage.Transaction, error) { return tx, nil }, func(storage.Transaction) error { return want })
	if !errors.Is(err, want) || tx.rollbacks != 1 || tx.commits != 0 {
		t.Fatalf("rollback contract broken: err=%v tx=%#v", err, tx)
	}
}

func TestCommitDoesNotSwallowBusinessError(t *testing.T) {
	businessErr := errors.New("module conflict")
	tx := &fakeTx{rollbackErr: errors.New("rollback transport failed")}
	err := storage.WithTransaction(context.Background(), func(context.Context) (storage.Transaction, error) { return tx, nil }, func(storage.Transaction) error { return businessErr })
	if !errors.Is(err, businessErr) || tx.commits != 0 || tx.rollbacks != 1 {
		t.Fatalf("business error was swallowed: err=%v tx=%#v", err, tx)
	}
}

type closeProbe struct{ closed *int }

func (c closeProbe) Close() error { *c.closed++; return nil }

func TestBatchResourcesClosePerItem(t *testing.T) {
	closed := 0
	m := &storage.Migrator{Operations: []storage.MigrationStep{
		{Name: "one", Apply: func(context.Context) (io.Closer, error) { return closeProbe{closed: &closed}, nil }},
		{Name: "two", Apply: func(context.Context) (io.Closer, error) {
			if closed != 1 {
				t.Fatalf("first resource still open before second step: %d", closed)
			}
			return closeProbe{closed: &closed}, nil
		}},
	}}
	if err := m.Run(context.Background()); err != nil || closed != 2 {
		t.Fatalf("migration cleanup failed: closed=%d err=%v", closed, err)
	}
}

func TestFailedBatchLeavesNoModules(t *testing.T) {
	repo := infrastructure.New()
	svc := moduleapp.New(repo)
	modules := []domain.Module{
		{Namespace: "acme", Name: "access", Version: "1.0.0"},
		{Namespace: "", Name: "broken", Version: "1.0.0"},
	}
	if err := svc.RegisterBatch(context.Background(), modules); err == nil {
		t.Fatal("invalid batch was accepted")
	}
	if repo.Exists("acme/access") {
		t.Fatal("failed batch left the first module behind")
	}
}
