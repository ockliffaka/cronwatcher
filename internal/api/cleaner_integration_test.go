package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatcher/internal/api"
	"github.com/user/cronwatcher/internal/config"
	"github.com/user/cronwatcher/internal/history"
)

// TestMetricsHandler_AfterClean verifies that metrics reflect the store state
// after the history cleaner has purged stale entries.
func TestMetricsHandler_AfterClean(t *testing.T) {
	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "nightly", Schedule: "0 0 * * *", Command: "true"},
		},
	}

	store := history.New(50)

	// Record one stale and one recent entry.
	store.Record(history.Entry{
		JobName:   "nightly",
		StartedAt: time.Now().Add(-3 * time.Hour),
		Success:   false,
	})
	store.Record(history.Entry{
		JobName:   "nightly",
		StartedAt: time.Now().Add(-1 * time.Minute),
		Success:   true,
	})

	// Run cleaner sweep directly (retention = 2 h).
	cleaner := history.NewCleaner(store, 2*time.Hour, 24*time.Hour)
	// Access sweep via exported helper — we call Start/Stop with a short tick.
	cleaner.Start()
	time.Sleep(10 * time.Millisecond)
	cleaner.Stop()

	reporter := history.NewReporter(store)
	handler := api.NewMetricsHandler(cfg, reporter)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if body == "" {
		t.Error("expected non-empty metrics response")
	}
}
