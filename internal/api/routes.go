package api

import (
	"net/http"

	"github.com/cronwatcher/internal/config"
	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/scheduler"
)

// Server bundles all HTTP handlers for cronwatcher's API.
type Server struct {
	mux *http.ServeMux
}

// NewServer wires up all routes and returns a ready-to-use Server.
func NewServer(
	cfg *config.Config,
	h *history.History,
	r *history.Reporter,
	sched *scheduler.Scheduler,
) *Server {
	mux := http.NewServeMux()

	statusHandler := NewStatusHandler(cfg, h, sched)
	metricsHandler := NewMetricsHandler(cfg, h, r)

	mux.Handle("/status", statusHandler)
	mux.Handle("/metrics", metricsHandler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &Server{mux: mux}
}

// Handler returns the underlying http.Handler for use with http.ListenAndServe.
func (s *Server) Handler() http.Handler {
	return s.mux
}
