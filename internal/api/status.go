package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/scheduler"
)

// StatusHandler serves a lightweight HTTP status endpoint that exposes
// the last run result and next scheduled run for each monitored job.
type StatusHandler struct {
	hist      *history.History
	sched     *scheduler.Scheduler
}

// NewStatusHandler creates a StatusHandler wired to the given history store
// and scheduler.
func NewStatusHandler(h *history.History, s *scheduler.Scheduler) *StatusHandler {
	return &StatusHandler{hist: h, sched: s}
}

type jobStatus struct {
	Job      string     `json:"job"`
	LastRun  *runResult `json:"last_run,omitempty"`
	NextRun  *time.Time `json:"next_run,omitempty"`
}

type runResult struct {
	StartedAt time.Time `json:"started_at"`
	Duration  string    `json:"duration"`
	Success   bool      `json:"success"`
	Output    string    `json:"output,omitempty"`
}

type statusResponse struct {
	GeneratedAt time.Time   `json:"generated_at"`
	Jobs        []jobStatus `json:"jobs"`
}

// ServeHTTP handles GET /status and returns JSON.
func (sh *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobNames := sh.sched.Jobs()
	statuses := make([]jobStatus, 0, len(jobNames))

	for _, name := range jobNames {
		js := jobStatus{Job: name}

		if entry, ok := sh.hist.Last(name); ok {
			js.LastRun = &runResult{
				StartedAt: entry.StartedAt,
				Duration:  entry.Duration.String(),
				Success:   entry.Success,
				Output:    entry.Output,
			}
		}

		if next, err := sh.sched.NextRun(name); err == nil {
			js.NextRun = &next
		}

		statuses = append(statuses, js)
	}

	resp := statusResponse{
		GeneratedAt: time.Now().UTC(),
		Jobs:        statuses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
