package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatcher/internal/throttle"
)

// NewThrottleHandler returns an HTTP handler for querying and resetting
// per-job execution throttle state.
//
//	GET  /throttle?job=<name>   — returns remaining executions in window
//	DELETE /throttle?job=<name> — resets the throttle record for the job
func NewThrottleHandler(store *throttle.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, `{"error":"missing job parameter"}`, http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			rem := store.Remaining(job)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"job":       job,
				"remaining": rem,
			})

		case http.MethodDelete:
			store.Reset(job)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "reset",
				"job":    job,
			})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}
