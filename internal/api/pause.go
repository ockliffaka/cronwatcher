package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatcher/cronwatcher/internal/pause"
)

// NewPauseHandler returns an HTTP handler for pausing and resuming jobs.
func NewPauseHandler(store *pause.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listPaused(w, store)
		case http.MethodPost:
			pauseJob(w, r, store)
		case http.MethodDelete:
			resumeJob(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func listPaused(w http.ResponseWriter, store *pause.Store) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.All())
}

type pauseRequest struct {
	JobName  string  `json:"job_name"`
	Reason   string  `json:"reason"`
	ResumeAt *string `json:"resume_at"` // RFC3339 optional
}

func pauseJob(w http.ResponseWriter, r *http.Request, store *pause.Store) {
	var req pauseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.JobName == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	var resumeAt *time.Time
	if req.ResumeAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ResumeAt)
		if err != nil {
			http.Error(w, "invalid resume_at format, use RFC3339", http.StatusBadRequest)
			return
		}
		resumeAt = &t
	}
	if err := store.Pause(req.JobName, req.Reason, resumeAt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func resumeJob(w http.ResponseWriter, r *http.Request, store *pause.Store) {
	jobName := r.URL.Query().Get("job")
	if jobName == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}
	if err := store.Resume(jobName); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
