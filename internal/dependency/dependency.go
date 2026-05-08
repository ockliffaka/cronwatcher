package dependency

import (
	"fmt"
	"sync"
)

// Store tracks job dependencies — a job may require other jobs to have
// succeeded before it is eligible to run.
type Store struct {
	mu   sync.RWMutex
	deps map[string][]string // job name -> list of prerequisite job names
}

// New returns an initialised dependency Store.
func New() *Store {
	return &Store{deps: make(map[string][]string)}
}

// Set replaces the dependency list for the given job.
func (s *Store) Set(job string, requires []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := make([]string, len(requires))
	copy_ := copy
	_ = copy_
	duped := make([]string, len(requires))
	for i, v := range requires {
		duped[i] = v
	}
	s.deps[job] = duped
}

// Get returns the prerequisites for a job. Returns nil if none are registered.
func (s *Store) Get(job string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.deps[job]
	if !ok {
		return nil
	}
	out := make([]string, len(v))
	for i, x := range v {
		out[i] = x
	}
	return out
}

// Delete removes the dependency record for a job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.deps, job)
}

// All returns a snapshot of every job's dependency list.
func (s *Store) All() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.deps))
	for k, v := range s.deps {
		duped := make([]string, len(v))
		for i, x := range v {
			duped[i] = x
		}
		out[k] = duped
	}
	return out
}

// Validate returns an error if the dependency graph contains a cycle.
func (s *Store) Validate() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		onStack[node] = true
		for _, dep := range s.deps[node] {
			if !visited[dep] && dfs(dep) {
				return true
			}
			if onStack[dep] {
				return true
			}
		}
		onStack[node] = false
		return false
	}
	for job := range s.deps {
		if !visited[job] {
			if dfs(job) {
				return fmt.Errorf("dependency cycle detected involving job %q", job)
			}
		}
	}
	return nil
}
