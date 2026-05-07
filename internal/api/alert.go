package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatcher/internal/alert"
)

type alertRequest struct {
	JobName string `json:"job_name"`
}

type alertResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// NewAlertHandler returns an HTTP handler that triggers alert evaluation for a job.
func NewAlertHandler(mgr *alert.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req alertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.JobName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(alertResponse{Status: "error", Message: "job_name is required"})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := mgr.Evaluate(req.JobName); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(alertResponse{Status: "error", Message: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(alertResponse{Status: "ok"})
	}
}
