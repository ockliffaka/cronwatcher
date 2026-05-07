package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_ReturnsOK(t *testing.T) {
	handler := NewHealthHandler("1.2.3")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}

	if resp.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", resp.Version)
	}

	if resp.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHealthHandler("0.0.1")

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		req := httptest.NewRequest(method, "/health", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s: expected 405, got %d", method, rec.Code)
		}
	}
}

func TestHealthHandler_ContentTypeJSON(t *testing.T) {
	handler := NewHealthHandler("dev")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
