package tag

import (
	"fmt"
	"sync"
)

// Store holds a mapping of job names to their associated tags.
type Store struct {
	mu   sync.RWMutex
	tags map[string][]string
}

// New creates an empty tag Store.
func New() *Store {
	return &Store{
		tags: make(map[string][]string),
	}
}

// Set replaces all tags for the given job.
func (s *Store) Set(job string, tags []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := make([]string, len(tags))
	for i, t := range tags {
		copy[i] = t
	}
	s.tags[job] = copy
}

// Get returns the tags associated with a job.
// Returns an error if the job is not found.
func (s *Store) Get(job string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tags[job]
	if !ok {
		return nil, fmt.Errorf("tag: job %q not found", job)
	}
	copy := make([]string, len(t))
	for i, v := range t {
		copy[i] = v
	}
	return copy, nil
}

// Delete removes all tags for a job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tags, job)
}

// All returns a snapshot of all job→tags mappings.
func (s *Store) All() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.tags))
	for k, v := range s.tags {
		cp := make([]string, len(v))
		for i, t := range v {
			cp[i] = t
		}
		out[k] = cp
	}
	return out
}

// HasTag reports whether the given job has the specified tag.
func (s *Store) HasTag(job, tag string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tags[job] {
		if t == tag {
			return true
		}
	}
	return false
}
