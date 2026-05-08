package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/cronwatcher/internal/audit"
)

// NewAuditHandler returns an HTTP handler that exposes the audit log.
// GET /api/audit          — returns all events
// GET /api/audit?kind=x   — filters by event kind
func NewAuditHandler(log *audit.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		kind := audit.EventKind(r.URL.Query().Get("kind"))
		events := log.Filter(kind)
		if events == nil {
			events = []audit.Event{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
