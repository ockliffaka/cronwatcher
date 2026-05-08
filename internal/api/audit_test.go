package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatcher/internal/audit"
)

func buildAuditHandler() (*audit.Log, http.HandlerFunc) {
	log := audit.New(50)
	return log, NewAuditHandler(log)
}

func TestAuditHandler_ReturnsAllEvents(t *testing.T) {
	log, h := buildAuditHandler()
	log.Record(audit.EventJobStarted, "backup", "started")
	log.Record(audit.EventAlertSent, "backup", "alert fired")

	req := httptest.NewRequest(http.MethodGet, "/api/audit", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []audit.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestAuditHandler_FiltersByKind(t *testing.T) {
	log, h := buildAuditHandler()
	log.Record(audit.EventJobStarted, "sync", "started")
	log.Record(audit.EventAlertSent, "sync", "alert")

	req := httptest.NewRequest(http.MethodGet, "/api/audit?kind=alert_sent", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	var events []audit.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 filtered event, got %d", len(events))
	}
}

func TestAuditHandler_MethodNotAllowed(t *testing.T) {
	_, h := buildAuditHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/audit", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAuditHandler_EmptyLog(t *testing.T) {
	_, h := buildAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/audit", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	var events []audit.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty list, got %d events", len(events))
	}
}
