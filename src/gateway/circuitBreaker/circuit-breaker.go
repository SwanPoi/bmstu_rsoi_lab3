package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

const (
	StateClosed = "closed"
	StateOpen = "open"
	StateHalfOpen = "half-open"
)

type CircuitBreaker struct {
	mu					sync.RWMutex
	State 				string
	FailuresQty			int
	MaxFailures			int
	LastFailureTime		time.Time
	ResetTimeout		time.Duration
	Timeout				time.Duration
}

func NewCircuitBreaker(maxFailures int, timeout, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		State:        StateClosed,
		MaxFailures:  maxFailures,
		Timeout:      timeout,
		ResetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.State {
		case StateClosed:
			return true
		case StateOpen:
			if time.Since(cb.LastFailureTime) >= cb.Timeout {
				cb.State = StateHalfOpen
				return true
			}
			return false
		case StateHalfOpen:
			return true
		default:
			return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.State == StateHalfOpen {
		cb.State = StateClosed
	}

	if time.Since(cb.LastFailureTime) >= cb.ResetTimeout {
		cb.FailuresQty = 0
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.FailuresQty++
	cb.LastFailureTime = time.Now()

	if cb.FailuresQty >= cb.MaxFailures {
		cb.State = StateOpen
	}

	if cb.State == StateHalfOpen {
		cb.State = StateOpen
	}
}

func (cb *CircuitBreaker) Execute(operation func() error, fallback func()) error {
    if !cb.AllowRequest() {
        fallback()
        return errors.New("circuit breaker open")
    }

    err := operation()
    if err != nil {
        cb.RecordFailure()
        fallback()
        return err
    }

    cb.RecordSuccess()
    return nil
}