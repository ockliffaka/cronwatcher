// Package runlock prevents concurrent execution of the same cron job.
package runlock

import (
	"fmt"
	"sync"
	"time"
)

// Store tracks which jobs are currently running.
type Store struct {
	mu      sync.Mutex
	running map[string]time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		running: make(map[string]time.Time),
	}
}

// Acquire marks a job as running. It returns an error if the job is already
// running. On success the caller must call Release when the job finishes.
func (s *Store) Acquire(jobName string) error {
	if jobName == "" {
		return fmt.Errorf("runlock: job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if start, ok := s.running[jobName]; ok {
		return fmt.Errorf("runlock: job %q is already running (started %s)", jobName, start.Format(time.RFC3339))
	}
	s.running[jobName] = time.Now()
	return nil
}

// Release removes the running lock for a job. It is a no-op if the job is not
// currently tracked.
func (s *Store) Release(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, jobName)
}

// IsRunning reports whether a job is currently executing.
func (s *Store) IsRunning(jobName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.running[jobName]
	return ok
}

// All returns a snapshot of all currently running jobs and their start times.
func (s *Store) All() map[string]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]time.Time, len(s.running))
	for k, v := range s.running {
		out[k] = v
	}
	return out
}
