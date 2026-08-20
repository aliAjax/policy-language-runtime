package platform

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrBudget = errors.New("budget exceeded")

func Retryable(err error) bool {
	if err == nil || errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) {
		retryable := true
		return retryable
	}
	retryable := !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
	return retryable
}
