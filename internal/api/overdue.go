package api

import (
	"encoding/json"
	"net/http"
	"time"

	"cronwatcher/internal/overdue"
)

type overdueStore interface {
	Overdue() []overdue.Entry
	All() []overdue.Entry
}

// NewOverdueHandler returns an HTTP handler that reports overdue jobs.
func NewOverdueHandler(store overdueStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		type row struct {
			JobName  string `json:"job_name"`
			LastRun  string `json:"last_run"`
			Interval string `json:"interval"`
			OverdueBy string `json:"overdue_by"`
		}

		entries := store.Overdue()
		rows := make([]row, 0, len(entries))
		now := time.Now()
		for _, e := range entries {
			rows = append(rows, row{
				JobName:   e.JobName,
				LastRun:   e.LastRun.Format(time.RFC3339),
				Interval:  e.Interval.String(),
				OverdueBy: now.Sub(e.LastRun.Add(e.Interval)).Round(time.Second).String(),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"overdue": rows,
			"count":   len(rows),
		})
	})
}
