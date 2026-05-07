package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cronwatcher/internal/api"
	"github.com/cronwatcher/internal/history"
)

func buildExportHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	store := history.New(10)
	store.Record(history.Entry{
		JobName:   "backup",
		Success:   true,
		StartedAt: time.Now().Add(-2 * time.Minute),
		Duration:  30 * time.Second,
	})
	store.Record(history.Entry{
		JobName:   "cleanup",
		Success:   false,
		StartedAt: time.Now().Add(-1 * time.Minute),
		Duration:  5 * time.Second,
		Error:     "exit status 1",
	})
	exporter := history.NewExporter(store)
	return api.NewExportHandler(exporter)
}

func TestExportHandler_JSON_Default(t *testing.T) {
	handler := buildExportHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
}

func TestExportHandler_CSV_Format(t *testing.T) {
	handler := buildExportHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/export?format=csv", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected text/csv, got %s", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "job_name") {
		t.Errorf("expected CSV header, got: %s", body)
	}
}

func TestExportHandler_UnsupportedFormat(t *testing.T) {
	handler := buildExportHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/export?format=xml", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestExportHandler_MethodNotAllowed(t *testing.T) {
	handler := buildExportHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/export", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
