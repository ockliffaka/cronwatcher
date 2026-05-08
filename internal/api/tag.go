package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/cronwatcher/internal/tag"
)

// NewTagHandler returns an HTTP handler for reading and writing job tags.
//
// GET  /api/tags?job=<name>  — retrieve tags for a job
// POST /api/tags             — set tags for a job (JSON body: {"job":"...","tags":[...]})
// DELETE /api/tags?job=<name> — remove all tags for a job
func NewTagHandler(store *tag.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			job := strings.TrimSpace(r.URL.Query().Get("job"))
			if job == "" {
				// Return all tags.
				json.NewEncoder(w).Encode(store.All())
				return
			}
			tags, err := store.Get(job)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"job": job, "tags": tags})

		case http.MethodPost:
			var body struct {
				Job  string   `json:"job"`
				Tags []string `json:"tags"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Job == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
				return
			}
			store.Set(body.Job, body.Tags)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			job := strings.TrimSpace(r.URL.Query().Get("job"))
			if job == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "job query param required"})
				return
			}
			store.Delete(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
