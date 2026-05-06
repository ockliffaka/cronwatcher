package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatcher/internal/config"
	"github.com/cronwatcher/internal/history"
)

func buildMetricsHandler(t *testing.T) (*MetricsHandler, *history.History) {
	t.Helper()
	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "backup", Schedule: "@daily", Command: "tar -czf /tmp/b.tgz /data"},
			{Name: "cleanup", Schedule: "@hourly", Command: "rm -rf /tmp/old"},
		},
	}
	h := history.New(10)
	r := history.NewReporter(h)
	return NewMetricsHandler(cfg, h, r), h
}

func TestMetricsHandler_ReturnsJSON(t *testing.T) {
	handler, h := buildMetricsHandler(t)

	h.Record("backup", history.Entry{Success: true, Duration: 2 * time.Second, RunAt: time.Now()})
	h.Record("backup", history.Entry{Success: false, Duration: 1 * time.Second, RunAt: time.Now()})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp MetricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	s, ok := resp.Jobs["backup"]
	if !ok {
		t.Fatal("expected 'backup' in response")
	}
	if s.Total != 2 {
		t.Errorf("expected Total=2, got %d", s.Total)
	}
	if s.Failures != 1 {
		t.Errorf("expected Failures=1, got %d", s.Failures)
	}
}

func TestMetricsHandler_MethodNotAllowed(t *testing.T) {
	handler, _ := buildMetricsHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestMetricsHandler_EmptyHistory(t *testing.T) {
	handler, _ := buildMetricsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp MetricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	// Jobs with no history are skipped by Summarise
	if len(resp.Jobs) != 0 {
		t.Errorf("expected empty jobs map, got %d entries", len(resp.Jobs))
	}
}
