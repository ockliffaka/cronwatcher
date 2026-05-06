package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatcher/internal/config"
	"github.com/cronwatcher/internal/history"
)

// MetricsResponse holds aggregated metrics for all configured jobs.
type MetricsResponse struct {
	Jobs map[string]history.Summary `json:"jobs"`
}

// MetricsHandler returns an HTTP handler that exposes per-job summary metrics.
type MetricsHandler struct {
	cfg      *config.Config
	history  *history.History
	reporter *history.Reporter
}

// NewMetricsHandler constructs a MetricsHandler.
func NewMetricsHandler(cfg *config.Config, h *history.History, r *history.Reporter) *MetricsHandler {
	return &MetricsHandler{cfg: cfg, history: h, reporter: r}
}

// ServeHTTP handles GET /metrics and returns JSON with per-job summaries.
func (m *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result := MetricsResponse{
		Jobs: make(map[string]history.Summary, len(m.cfg.Jobs)),
	}

	for _, job := range m.cfg.Jobs {
		summary, err := m.reporter.Summarise(job.Name)
		if err != nil {
			continue
		}
		result.Jobs[job.Name] = summary
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
