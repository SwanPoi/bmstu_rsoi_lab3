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

// TODO: переделать под циклический массив
type CircuitBreaker struct {
	mu					sync.RWMutex
	State 				string
	RingBuffer			*RingBuffer
	FailureRate			float64
	LastRequestTime		time.Time
	Timeout				time.Duration
}

func NewCircuitBreaker(bufferSize int, failureRate float64, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		State:        	StateClosed,
		RingBuffer: 	NewRingBuffer(bufferSize),
		FailureRate: 	failureRate,
		Timeout:      	timeout,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.State {
		case StateClosed:
			return true
		case StateOpen:
			if time.Since(cb.LastRequestTime) >= cb.Timeout {
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

	cb.RingBuffer.Add(RequestResult{Success: true, Time: time.Now()})

	if cb.State == StateHalfOpen {
		cb.State = StateClosed
	}

	if cb.State == StateClosed && cb.RingBuffer.GetFailureRate() <= cb.FailureRate {
	} else if cb.State == StateClosed && cb.RingBuffer.GetFailureRate() > cb.FailureRate {
		cb.State = StateOpen
		cb.LastRequestTime = time.Now()
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.RingBuffer.Add(RequestResult{Success: false, Time: time.Now()})
	
	if cb.State == StateHalfOpen {
		cb.State = StateOpen
		cb.LastRequestTime = time.Now()
		return
	}
	
	if cb.State == StateClosed && cb.RingBuffer.GetFailureRate() > cb.FailureRate {
		cb.State = StateOpen
		cb.LastRequestTime = time.Now()
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

func (cb *CircuitBreaker) GetFailureRate() float64 {
	return cb.RingBuffer.GetFailureRate()
}