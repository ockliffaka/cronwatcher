package ratelimit

import (
	"sync"
	"time"
)

// Limiter prevents alert flooding by suppressing repeated alerts
// for the same job within a cooldown window.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
}

// New creates a Limiter with the given cooldown duration.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
	}
}

// Allow returns true if an alert for jobName is permitted, i.e. no alert
// has been sent within the cooldown window. Calling Allow records the
// current time when it returns true.
func (l *Limiter) Allow(jobName string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if t, ok := l.last[jobName]; ok {
		if now.Sub(t) < l.cooldown {
			return false
		}
	}
	l.last[jobName] = now
	return true
}

// Reset clears the rate-limit record for jobName, allowing the next
// alert through immediately.
func (l *Limiter) Reset(jobName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, jobName)
}

// LastSent returns the time of the most recent allowed alert for
// jobName and whether a record exists.
func (l *Limiter) LastSent(jobName string) (time.Time, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	t, ok := l.last[jobName]
	return t, ok
}
