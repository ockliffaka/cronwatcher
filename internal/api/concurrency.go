package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/example/cronwatcher/internal/concurrency"
)

// NewConcurrencyHandler returns an http.Handler that exposes concurrency state.
//
// GET  /api/concurrency         — list all entries
// GET  /api/concurrency?job=X   — get entry for a single job
// POST /api/concurrency         — set max for a job  {"job":"x","max":2}
// POST /api/concurrency/release — release one slot   {"job":"x"}
func NewConcurrencyHandler(store *concurrency.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Distinguish /api/concurrency/release from /api/concurrency
		if strings.HasSuffix(r.URL.Path, "/release") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
				return
			}
			var body struct {
				Job string `json:"job"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Job == "" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing job"})
				return
			}
			store.Release(body.Job)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if job := r.URL.Query().Get("job"); job != "" {
				e, ok := store.Get(job)
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "job not found"})
					return
				}
				_ = json.NewEncoder(w).Encode(e)
				return
			}
			_ = json.NewEncoder(w).Encode(store.All())

		case http.MethodPost:
			var body struct {
				Job string `json:"job"`
				Max int    `json:"max"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Job == "" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing job"})
				return
			}
			if err := store.SetMax(body.Job, body.Max); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		}
	})
}
