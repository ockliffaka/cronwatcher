package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatcher/internal/dependency"
)

// NewDependencyHandler returns an http.Handler that exposes the dependency
// store over HTTP.
//
//   GET  /api/dependencies?job=<name>  — list deps for a single job
//   GET  /api/dependencies             — list all deps
//   POST /api/dependencies             — set deps for a job (JSON body)
//   DELETE /api/dependencies?job=<name> — remove deps for a job
func NewDependencyHandler(store *dependency.Store) http.HandlerFunc {
	type setRequest struct {
		Job      string   `json:"job"`
		Requires []string `json:"requires"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			job := r.URL.Query().Get("job")
			if job != "" {
				deps := store.Get(job)
				if deps == nil {
					deps = []string{}
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"job":      job,
					"requires": deps,
				})
				return
			}
			json.NewEncoder(w).Encode(store.All())

		case http.MethodPost:
			var req setRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			store.Set(req.Job, req.Requires)
			if err := store.Validate(); err != nil {
				store.Delete(req.Job)
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnprocessableEntity)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, `{"error":"missing job query param"}`, http.StatusBadRequest)
				return
			}
			store.Delete(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}
