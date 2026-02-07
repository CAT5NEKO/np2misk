package resilience

import (
	"log"
	"sync"
	"time"

	"np2misk/internal/domain/repository"
)

type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half-open"
)

type CircuitBreakerConfig struct {
	MaxFailures  int
	ResetTimeout time.Duration
}

type circuitBreaker struct {
	mu               sync.Mutex
	failures         int
	maxFailures      int
	resetTimeout     time.Duration
	lastFailure      time.Time
	state            CircuitBreakerState
	halfOpenAttempts int
}

func NewCircuitBreaker(cfg CircuitBreakerConfig) repository.CircuitBreaker {
	maxFailures := cfg.MaxFailures
	if maxFailures == 0 {
		maxFailures = 5
	}
	resetTimeout := cfg.ResetTimeout
	if resetTimeout == 0 {
		resetTimeout = 60 * time.Second
	}
	return &circuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

func (cb *circuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenAttempts = 0
			return true
		}
		return false
	case StateHalfOpen:
		cb.halfOpenAttempts++
		return cb.halfOpenAttempts <= 1
	default:
		return true
	}
}

func (cb *circuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

func (cb *circuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		log.Printf("Circuit breaker opened after %d failures", cb.failures)
	}
}
