package resilience

import (
	"context"
	"log"
	"time"

	"np2misk/internal/domain/repository"
)

type RetryerConfig struct {
	MaxRetries int
	Delay      time.Duration
}

type retryer struct {
	maxRetries int
	delay      time.Duration
}

func NewRetryer(cfg RetryerConfig) repository.Retryer {
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}
	delay := cfg.Delay
	if delay == 0 {
		delay = 2 * time.Second
	}
	return &retryer{
		maxRetries: maxRetries,
		delay:      delay,
	}
}

func (r *retryer) Execute(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := range r.maxRetries + 1 {
		if i > 0 {
			log.Printf("Retry attempt %d/%d", i, r.maxRetries)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.delay):
			}
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("Operation failed: %v", err)
		}
	}
	return lastErr
}
