package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/config"
	"github.com/cronwatcher/internal/scheduler"
)

func baseConfig(jobs []config.Job) *config.Config {
	return &config.Config{Jobs: jobs}
}

func TestStatusHandler_ReturnsJSON(t *testing.T) {
	cfg := baseConfig([]config.Job{
		{Name: "backup", Schedule: "@hourly", Command: "echo ok"},
	})

	hist := history.New(cfg, 10)
	hist.Record(history.Entry{
		Job:       "backup",
		StartedAt: time.Now().UTC(),
		Duration:  2 * time.Second,
		Success:   true,
		Output:    "ok",
	})

	sched, _ := scheduler.New(cfg, nil, nil)

	handler := NewStatusHandler(hist, sched)

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(resp.Jobs))
	}
	if resp.Jobs[0].Job != "backup" {
		t.Errorf("expected job name 'backup', got %q", resp.Jobs[0].Job)
	}
	if resp.Jobs[0].LastRun == nil {
		t.Error("expected last_run to be populated")
	} else if !resp.Jobs[0].LastRun.Success {
		t.Error("expected last_run.success to be true")
	}
}

func TestStatusHandler_MethodNotAllowed(t *testing.T) {
	cfg := baseConfig([]config.Job{})
	hist := history.New(cfg, 10)
	sched, _ := scheduler.New(cfg, nil, nil)

	handler := NewStatusHandler(hist, sched)

	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestStatusHandler_NoLastRun(t *testing.T) {
	cfg := baseConfig([]config.Job{
		{Name: "cleanup", Schedule: "@daily", Command: "rm -rf /tmp/old"},
	})
	hist := history.New(cfg, 10)
	sched, _ := scheduler.New(cfg, nil, nil)

	handler := NewStatusHandler(hist, sched)

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp statusResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Jobs[0].LastRun != nil {
		t.Error("expected last_run to be nil for a job that has never run")
	}
}
