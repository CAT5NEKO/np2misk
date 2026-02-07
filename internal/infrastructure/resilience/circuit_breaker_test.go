package resilience

import (
	"testing"
	"time"
)

func TestCircuitBreaker_AllowWhenClosed(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
	})

	if !cb.Allow() {
		t.Error("expected Allow() to return true when closed")
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
	})

	for range 3 {
		cb.RecordFailure()
	}

	if cb.Allow() {
		t.Error("expected Allow() to return false when open")
	}
}

func TestCircuitBreaker_ResetsAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxFailures:  2,
		ResetTimeout: 50 * time.Millisecond,
	})

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.Allow() {
		t.Error("expected circuit to be open")
	}

	time.Sleep(60 * time.Millisecond)

	if !cb.Allow() {
		t.Error("expected circuit to be half-open after timeout")
	}

	cb.RecordSuccess()

	if !cb.Allow() {
		t.Error("expected circuit to be closed after success")
	}
}

func TestCircuitBreaker_DefaultValues(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	for range 5 {
		cb.RecordFailure()
	}

	if cb.Allow() {
		t.Error("expected default maxFailures to be 5")
	}
}
