package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/cronwatcher/internal/config"
	"github.com/cronwatcher/internal/history"
)

// RunWithTimeout executes a job and enforces a maximum duration.
// If the job exceeds its configured timeout, the context is cancelled and
// a failure entry is recorded in the history store.
func RunWithTimeout(cfg *config.Config, jobName string, store *history.Store) error {
	job, ok := findJob(cfg, jobName)
	if !ok {
		return fmt.Errorf("job %q not found", jobName)
	}

	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	err := runJobContext(ctx, job)
	duration := time.Since(start)

	entry := history.Entry{
		JobName:   jobName,
		StartedAt: start,
		Duration:  duration,
		Success:   err == nil,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	store.Record(entry)

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("job %q exceeded timeout of %s", jobName, timeout)
	}
	return err
}

func findJob(cfg *config.Config, name string) (config.Job, bool) {
	for _, j := range cfg.Jobs {
		if j.Name == name {
			return j, true
		}
	}
	return config.Job{}, false
}
