package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatcher/internal/runlock"
)

type runlockEntry struct {
	Job       string    `json:"job"`
	StartedAt time.Time `json:"started_at"`
}

// NewRunlockHandler returns an HTTP handler that exposes which jobs are
// currently executing according to the provided runlock Store.
//
//	GET /api/runlock  — list all running jobs
func NewRunlockHandler(store *runlock.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		all := store.All()
		entries := make([]runlockEntry, 0, len(all))
		for job, start := range all {
			entries = append(entries, runlockEntry{Job: job, StartedAt: start})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}
