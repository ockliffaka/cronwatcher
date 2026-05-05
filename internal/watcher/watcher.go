package watcher

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/user/cronwatcher/internal/config"
	"github.com/user/cronwatcher/internal/notify"
)

// JobResult holds the outcome of a single cron job execution.
type JobResult struct {
	JobName  string
	Command  string
	ExitCode int
	Duration time.Duration
	Err      error
}

// Runner executes jobs and reports failures via a notifier.
type Runner struct {
	cfg      *config.Config
	notifier notify.Notifier
}

// New creates a Runner with the given config and notifier.
func New(cfg *config.Config, n notify.Notifier) *Runner {
	return &Runner{cfg: cfg, notifier: n}
}

// RunJob executes a single job by name and returns its result.
func (r *Runner) RunJob(name string) JobResult {
	var job *config.Job
	for i := range r.cfg.Jobs {
		if r.cfg.Jobs[i].Name == name {
			job = &r.cfg.Jobs[i]
			break
		}
	}
	if job == nil {
		return JobResult{JobName: name, Err: fmt.Errorf("job %q not found", name)}
	}

	start := time.Now()
	cmd := exec.Command("sh", "-c", job.Command)
	err := cmd.Run()
	duration := time.Since(start)

	result := JobResult{
		JobName:  job.Name,
		Command:  job.Command,
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Err = err
		_ = r.notifier.Send(fmt.Sprintf(
			"[cronwatcher] job %q FAILED (exit %d) after %s: %v",
			job.Name, result.ExitCode, duration.Round(time.Millisecond), err,
		))
	}

	return result
}
