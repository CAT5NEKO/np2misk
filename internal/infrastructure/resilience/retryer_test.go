package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryer_SuccessOnFirstAttempt(t *testing.T) {
	r := NewRetryer(RetryerConfig{MaxRetries: 3, Delay: time.Millisecond})

	callCount := 0
	err := r.Execute(context.Background(), func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1", callCount)
	}
}

func TestRetryer_RetriesOnFailure(t *testing.T) {
	r := NewRetryer(RetryerConfig{MaxRetries: 3, Delay: time.Millisecond})

	callCount := 0
	testErr := errors.New("test error")
	err := r.Execute(context.Background(), func() error {
		callCount++
		return testErr
	})

	if err == nil {
		t.Error("expected error but got nil")
	}
	if callCount != 4 {
		t.Errorf("callCount = %d, want 4", callCount)
	}
}

func TestRetryer_SuccessAfterRetries(t *testing.T) {
	r := NewRetryer(RetryerConfig{MaxRetries: 3, Delay: time.Millisecond})

	callCount := 0
	err := r.Execute(context.Background(), func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestRetryer_RespectsContext(t *testing.T) {
	r := NewRetryer(RetryerConfig{MaxRetries: 10, Delay: 100 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := r.Execute(ctx, func() error {
		callCount++
		return errors.New("error")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetryer_DefaultValues(t *testing.T) {
	r := NewRetryer(RetryerConfig{})

	callCount := 0
	_ = r.Execute(context.Background(), func() error {
		callCount++
		return errors.New("error")
	})

	if callCount != 4 {
		t.Errorf("expected default maxRetries to be 3, callCount = %d", callCount)
	}
}
