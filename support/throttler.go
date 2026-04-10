package support

import (
	"sync"
	"time"
)

const DefaultThrottleInterval = 10 * time.Second

type Throttler[T comparable] struct {
	mu       sync.RWMutex
	interval time.Duration
	lastAt   map[T]time.Time
}

func NewThrottler[T comparable](interval time.Duration) *Throttler[T] {
	return &Throttler[T]{
		interval: interval,
		lastAt:   make(map[T]time.Time),
	}
}

func (t *Throttler[T]) ShouldThrottle(key T) bool {
	now := time.Now()

	t.mu.RLock()
	lastAt, exists := t.lastAt[key]
	if exists && now.Sub(lastAt) < t.interval {
		t.mu.RUnlock()
		return true
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()
	// Double check under write lock
	lastAt, exists = t.lastAt[key]
	if exists && now.Sub(lastAt) < t.interval {
		return true
	}

	t.lastAt[key] = now
	return false
}
