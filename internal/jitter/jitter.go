// Package jitter provides utilities for adding randomised delay to cron job
// execution, preventing thundering-herd problems when many jobs share the same
// schedule.
package jitter

import (
	"math/rand"
	"sync"
	"time"
)

// Store holds per-job jitter configuration.
type Store struct {
	mu      sync.RWMutex
	entries map[string]time.Duration // max jitter per job
	rng     *rand.Rand
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]time.Duration),
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
	}
}

// Set registers a maximum jitter duration for the named job.
// A zero or negative max is silently ignored.
func (s *Store) Set(job string, max time.Duration) error {
	if job == "" {
		return ErrEmptyJobName
	}
	if max <= 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = max
	return nil
}

// Delete removes jitter configuration for the named job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// Delay returns a random duration in [0, max) for the named job.
// If the job has no registered jitter, zero is returned.
func (s *Store) Delay(job string) time.Duration {
	s.mu.RLock()
	max, ok := s.entries[job]
	s.mu.RUnlock()
	if !ok || max <= 0 {
		return 0
	}
	s.mu.Lock()
	delay := time.Duration(s.rng.Int63n(int64(max)))
	s.mu.Unlock()
	return delay
}

// All returns a snapshot of all registered jitter entries.
func (s *Store) All() map[string]time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]time.Duration, len(s.entries))
	for k, v := range s.entries {
		out[k] = v
	}
	return out
}

// ErrEmptyJobName is returned when an empty job name is supplied.
var ErrEmptyJobName = errEmptyJobName("job name must not be empty")

type errEmptyJobName string

func (e errEmptyJobName) Error() string { return string(e) }
