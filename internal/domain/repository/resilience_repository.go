package repository

import "context"

type CircuitBreaker interface {
	Allow() bool
	RecordSuccess()
	RecordFailure()
}

type Retryer interface {
	Execute(ctx context.Context, fn func() error) error
}
