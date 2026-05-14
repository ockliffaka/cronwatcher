package throttle

import (
	"errors"
	"sync"
	"time"
)

// Entry tracks execution count and window start for a single job.
type Entry struct {
	count     int
	windowStart time.Time
}

// Store enforces a maximum number of executions per job within a rolling window.
type Store struct {
	mu       sync.Mutex
	records  map[string]*Entry
	max      int
	window   time.Duration
}

// New creates a Store that allows at most max executions per window.
func New(max int, window time.Duration) (*Store, error) {
	if max <= 0 {
		return nil, errors.New("throttle: max must be greater than zero")
	}
	if window <= 0 {
		return nil, errors.New("throttle: window must be greater than zero")
	}
	return &Store{
		records: make(map[string]*Entry),
		max:     max,
		window:  window,
	}, nil
}

// Allow returns true if the job may execute, recording the attempt.
// Returns false when the job has reached the execution limit within the window.
func (s *Store) Allow(job string) bool {
	if job == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	e, ok := s.records[job]
	if !ok || now.Sub(e.windowStart) >= s.window {
		s.records[job] = &Entry{count: 1, windowStart: now}
		return true
	}
	if e.count >= s.max {
		return false
	}
	e.count++
	return true
}

// Remaining returns how many executions are left in the current window.
func (s *Store) Remaining(job string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.records[job]
	if !ok || time.Since(e.windowStart) >= s.window {
		return s.max
	}
	rem := s.max - e.count
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the throttle record for a job.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, job)
}
