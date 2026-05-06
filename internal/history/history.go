package history

import (
	"sync"
	"time"
)

// Status represents the result of a cron job execution.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
)

// Entry records the outcome of a single job run.
type Entry struct {
	JobName   string
	Status    Status
	Output    string
	Error     string
	StartedAt time.Time
	Duration  time.Duration
}

// Store holds an in-memory history of job executions.
type Store struct {
	mu      sync.RWMutex
	entries map[string][]Entry
	maxPer  int
}

// New creates a new Store keeping at most maxPerJob entries per job.
func New(maxPerJob int) *Store {
	if maxPerJob <= 0 {
		maxPerJob = 50
	}
	return &Store{
		entries: make(map[string][]Entry),
		maxPer:  maxPerJob,
	}
}

// Record appends an entry for the given job, evicting the oldest if needed.
func (s *Store) Record(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.entries[e.JobName]
	list = append(list, e)
	if len(list) > s.maxPer {
		list = list[len(list)-s.maxPer:]
	}
	s.entries[e.JobName] = list
}

// Get returns a copy of all recorded entries for a job.
func (s *Store) Get(jobName string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	src := s.entries[jobName]
	out := make([]Entry, len(src))
	copy(out, src)
	return out
}

// Last returns the most recent entry for a job and whether one exists.
func (s *Store) Last(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := s.entries[jobName]
	if len(list) == 0 {
		return Entry{}, false
	}
	return list[len(list)-1], true
}
