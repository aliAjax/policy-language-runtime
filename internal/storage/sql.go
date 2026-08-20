package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type SQL struct{ DB *sql.DB }

func (s SQL) Ping(ctx context.Context) error {
	if s.DB == nil {
		return nil
	}
	return s.DB.PingContext(ctx)
}
func (s SQL) Exec(ctx context.Context, q string, args ...any) error {
	if s.DB == nil {
		return nil
	}
	_, e := s.DB.ExecContext(ctx, q, args...)
	return e
}

type Transaction interface {
	Commit() error
	Rollback() error
}

type BeginFunc func(context.Context) (Transaction, error)

func WithTransaction(ctx context.Context, begin BeginFunc, work func(Transaction) error) (err error) {
	tx, err := begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("begin transaction: nil transaction")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, fmt.Errorf("rollback transaction: %w", rollbackErr))
			}
		}
		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("commit transaction: %w", commitErr)
		}
	}()
	err = work(tx)
	return err
}
