package pause

import (
	"fmt"
	"sync"
	"time"
)

// Store tracks paused cron jobs and their optional resume times.
type Store struct {
	mu      sync.RWMutex
	paused  map[string]*Entry
}

// Entry holds pause metadata for a single job.
type Entry struct {
	JobName   string    `json:"job_name"`
	PausedAt  time.Time `json:"paused_at"`
	ResumeAt  *time.Time `json:"resume_at,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}

// New creates a new pause Store.
func New() *Store {
	return &Store{paused: make(map[string]*Entry)}
}

// Pause marks a job as paused. resumeAt may be nil for indefinite pause.
func (s *Store) Pause(jobName, reason string, resumeAt *time.Time) error {
	if jobName == "" {
		return fmt.Errorf("job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused[jobName] = &Entry{
		JobName:  jobName,
		PausedAt: time.Now().UTC(),
		ResumeAt: resumeAt,
		Reason:   reason,
	}
	return nil
}

// Resume removes a job from the paused set.
func (s *Store) Resume(jobName string) error {
	if jobName == "" {
		return fmt.Errorf("job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.paused[jobName]; !ok {
		return fmt.Errorf("job %q is not paused", jobName)
	}
	delete(s.paused, jobName)
	return nil
}

// IsPaused returns true when the job is currently paused.
// It also auto-resumes jobs whose resumeAt time has passed.
func (s *Store) IsPaused(jobName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.paused[jobName]
	if !ok {
		return false
	}
	if e.ResumeAt != nil && time.Now().UTC().After(*e.ResumeAt) {
		delete(s.paused, jobName)
		return false
	}
	return true
}

// All returns a copy of all paused entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.paused))
	for _, e := range s.paused {
		out = append(out, *e)
	}
	return out
}
