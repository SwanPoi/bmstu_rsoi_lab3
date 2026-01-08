package circuitbreaker

import (
	"sync"
	"time"
)

type RequestResult struct {
	Success 	bool
	Time		time.Time
}

type RingBuffer struct {
	buffer 	[]RequestResult
	size 	int
	head 	int
	count 	int
	mu 		sync.RWMutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer: 	make([]RequestResult, size),
		size: 		size,
		head: 		0,
		count: 		0,
	}
}

func (rb *RingBuffer) Add(result RequestResult) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.head] = result
	rb.head = (rb.head + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	}
}

func (rb *RingBuffer) GetFailureRate() float64 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return 0.0
	}

	failures := 0
	for i := 0; i < rb.count; i++ {
		idx := (rb.head - 1 - i + rb.size) % rb.size

		if !rb.buffer[idx].Success {
			failures++
		}
	}

	return float64(failures) / float64(rb.count)
}

func (rb *RingBuffer) IsRecentFailure() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	
	if rb.count == 0 {
		return false
	}
	
	lastIdx := (rb.head - 1 + rb.size) % rb.size
	return !rb.buffer[lastIdx].Success
}

func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	rb.head = 0
	rb.count = 0
}