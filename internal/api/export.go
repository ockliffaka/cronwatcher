package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cronwatcher/internal/history"
)

// NewExportHandler returns an HTTP handler that streams job history exports.
func NewExportHandler(exp *history.Exporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		format := strings.ToLower(r.URL.Query().Get("format"))
		if format == "" {
			format = "json"
		}

		switch format {
		case "json":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", "attachment; filename=\"history.json\"")
		case "csv":
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment; filename=\"history.csv\"")
		default:
			http.Error(w, fmt.Sprintf("unsupported format: %s", format), http.StatusBadRequest)
			return
		}

		if err := exp.Write(w, format); err != nil {
			http.Error(w, "failed to export history", http.StatusInternalServerError)
			return
		}
	}
}
