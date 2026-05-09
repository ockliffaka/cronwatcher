package api

import (
	"net/http"

	"github.com/cronwatcher/cronwatcher/internal/alert"
	"github.com/cronwatcher/cronwatcher/internal/audit"
	"github.com/cronwatcher/cronwatcher/internal/dependency"
	"github.com/cronwatcher/cronwatcher/internal/history"
	"github.com/cronwatcher/cronwatcher/internal/pause"
	"github.com/cronwatcher/cronwatcher/internal/tag"
)

// ServerDeps holds all dependencies needed to build the HTTP server.
type ServerDeps struct {
	History    *history.Store
	Reporter   *history.Reporter
	Exporter   *history.Exporter
	Alerts     *alert.Manager
	AuditLog   *audit.Log
	Tags       *tag.Store
	Deps       *dependency.Store
	Pauses     *pause.Store
}

// NewServer constructs and returns the root HTTP mux with all routes registered.
func NewServer(d ServerDeps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/health", NewHealthHandler())
	mux.Handle("/status", NewStatusHandler(d.History, d.Reporter))
	mux.Handle("/metrics", NewMetricsHandler(d.History, d.Reporter))
	mux.Handle("/export", NewExportHandler(d.Exporter))
	mux.Handle("/alert", NewAlertHandler(d.Alerts))
	mux.Handle("/audit", audit.Middleware(d.AuditLog, NewAuditHandler(d.AuditLog)))
	mux.Handle("/tags", NewTagHandler(d.Tags))
	mux.Handle("/dependencies", NewDependencyHandler(d.Deps))
	mux.Handle("/pause", NewPauseHandler(d.Pauses))

	return mux
}
