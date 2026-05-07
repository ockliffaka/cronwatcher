package api

import (
	"net/http"

	"github.com/cronwatcher/internal/alert"
	"github.com/cronwatcher/internal/config"
	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/watcher"
)

// ServerDeps groups all dependencies required to build the API server.
type ServerDeps struct {
	Config   *config.Config
	Store    *history.Store
	Watcher  *watcher.Watcher
	Reporter *history.Reporter
	Exporter *history.Exporter
	Alert    *alert.Manager
}

// NewServer wires up all HTTP routes and returns a ready-to-use http.ServeMux.
func NewServer(deps ServerDeps) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", NewStatusHandler(deps.Config, deps.Store, deps.Watcher))
	mux.HandleFunc("/metrics", NewMetricsHandler(deps.Store, deps.Reporter))
	mux.HandleFunc("/alert", NewAlertHandler(deps.Alert, deps.Config))
	mux.HandleFunc("/export", NewExportHandler(deps.Exporter))

	return mux
}
