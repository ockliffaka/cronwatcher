package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

// Breaker tracks consecutive failures per job and opens the circuit
// after a configurable threshold, preventing further executions until
// a cooldown period has elapsed.
type Breaker struct {
	mu           sync.Mutex
	failures     map[string]int
	openedAt     map[string]time.Time
	threshold    int
	cooldown     time.Duration
}

// New creates a Breaker with the given failure threshold and cooldown.
func New(threshold int, cooldown time.Duration) *Breaker {
	return &Breaker{
		failures:  make(map[string]int),
		openedAt:  make(map[string]time.Time),
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// Allow returns nil if the job may proceed, or ErrCircuitOpen if the
// circuit is open and the cooldown has not yet elapsed.
func (b *Breaker) Allow(job string) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	if opened, ok := b.openedAt[job]; ok {
		if time.Since(opened) < b.cooldown {
			return ErrCircuitOpen
		}
		// Cooldown elapsed — move to half-open: clear opened timestamp.
		delete(b.openedAt, job)
	}
	return nil
}

// RecordSuccess resets the failure counter for the job.
func (b *Breaker) RecordSuccess(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.failures, job)
	delete(b.openedAt, job)
}

// RecordFailure increments the failure counter and opens the circuit
// if the threshold is reached.
func (b *Breaker) RecordFailure(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures[job]++
	if b.failures[job] >= b.threshold {
		if _, alreadyOpen := b.openedAt[job]; !alreadyOpen {
			b.openedAt[job] = time.Now()
		}
	}
}

// State returns the current circuit state for a job.
func (b *Breaker) State(job string) State {
	b.mu.Lock()
	defer b.mu.Unlock()
	opened, isOpen := b.openedAt[job]
	if !isOpen {
		return StateClosed
	}
	if time.Since(opened) >= b.cooldown {
		return StateHalfOpen
	}
	return StateOpen
}

// Reset clears all state for a job.
func (b *Breaker) Reset(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.failures, job)
	delete(b.openedAt, job)
}
