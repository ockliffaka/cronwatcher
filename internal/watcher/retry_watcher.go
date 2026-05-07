package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/retry"
)

// RunJobWithRetry executes a named job with the given retry policy.
// Each attempt is recorded in the history store. Only the final outcome
// is treated as the canonical result for alerting purposes.
func (w *Watcher) RunJobWithRetry(ctx context.Context, jobName string, p retry.Policy) error {
	job, ok := w.jobs[jobName]
	if !ok {
		return fmt.Errorf("job %q not found", jobName)
	}

	var lastEntry history.Entry

	res := retry.Do(ctx, p, func() error {
		start := time.Now()
		err := w.exec(job.Command)
		dur := time.Since(start)

		status := "success"
		msg := ""
		if err != nil {
			status = "failure"
			msg = err.Error()
		}

		lastEntry = history.Entry{
			JobName:   jobName,
			Status:    status,
			Message:   msg,
			Timestamp: time.Now(),
			Duration:  dur,
		}
		return err
	})

	w.store.Record(lastEntry)
	return res.Err
}
