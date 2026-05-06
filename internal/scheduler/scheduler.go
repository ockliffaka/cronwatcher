package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"cronwatcher/internal/config"
	"cronwatcher/internal/watcher"
)

// Scheduler manages cron job scheduling and execution.
type Scheduler struct {
	cron    *cron.Cron
	watcher *watcher.Watcher
	cfg     *config.Config
	mu      sync.Mutex
	entries map[string]cron.EntryID
}

// New creates a new Scheduler with the given config and watcher.
func New(cfg *config.Config, w *watcher.Watcher) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		watcher: w,
		cfg:     cfg,
		entries: make(map[string]cron.EntryID),
	}
}

// Start registers all jobs from config and starts the cron scheduler.
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.cfg.Jobs {
		job := job // capture loop variable
		id, err := s.cron.AddFunc(job.Schedule, func() {
			log.Printf("[scheduler] running job: %s", job.Name)
			if err := s.watcher.RunJob(job.Name); err != nil {
				log.Printf("[scheduler] job %s failed: %v", job.Name, err)
			}
		})
		if err != nil {
			return err
		}
		s.entries[job.Name] = id
		log.Printf("[scheduler] registered job %q with schedule %q", job.Name, job.Schedule)
	}

	s.cron.Start()
	log.Println("[scheduler] started")
	return nil
}

// Stop gracefully stops the scheduler and waits for running jobs to finish.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
		log.Println("[scheduler] stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Println("[scheduler] stop timed out")
	}
}

// NextRun returns the next scheduled time for a named job.
func (s *Scheduler) NextRun(jobName string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.entries[jobName]
	if !ok {
		return time.Time{}, false
	}
	entry := s.cron.Entry(id)
	return entry.Next, true
}
