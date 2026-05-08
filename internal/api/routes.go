package api

import (
	"net/http"

	"github.com/user/cronwatcher/internal/alert"
	"github.com/user/cronwatcher/internal/audit"
	"github.com/user/cronwatcher/internal/history"
	"github.com/user/cronwatcher/internal/tag"
)

// ServerDeps holds all dependencies required to build the HTTP server.
type ServerDeps struct {
	History  *history.Store
	Reporter *history.Reporter
	Exporter *history.Exporter
	Alert    *alert.Manager
	Audit    *audit.Log
	Tags     *tag.Store
}

// NewServer wires all HTTP handlers onto a new ServeMux and returns it.
func NewServer(deps ServerDeps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/health", NewHealthHandler())
	mux.Handle("/api/status", NewStatusHandler(deps.History, deps.Reporter))
	mux.Handle("/api/metrics", NewMetricsHandler(deps.History, deps.Reporter))
	mux.Handle("/api/alert", NewAlertHandler(deps.Alert))
	mux.Handle("/api/export", NewExportHandler(deps.Exporter))
	mux.Handle("/api/audit", NewAuditHandler(deps.Audit))
	mux.Handle("/api/tags", NewTagHandler(deps.Tags))

	return audit.Middleware(deps.Audit, mux)
}
