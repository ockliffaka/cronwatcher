package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatcher/cronwatcher/internal/api"
	"github.com/cronwatcher/cronwatcher/internal/pause"
)

func buildPauseHandler() (http.Handler, *pause.Store) {
	store := pause.New()
	return api.NewPauseHandler(store), store
}

func TestPauseHandler_PauseAndList(t *testing.T) {
	h, _ := buildPauseHandler()
	body := `{"job_name":"backup","reason":"maintenance"}`
	req := httptest.NewRequest(http.MethodPost, "/pause", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/pause", nil)
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	var entries []map[string]interface{}
	json.NewDecoder(rr2.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0]["job_name"] != "backup" {
		t.Errorf("unexpected job name: %v", entries[0]["job_name"])
	}
}

func TestPauseHandler_Resume(t *testing.T) {
	h, store := buildPauseHandler()
	_ = store.Pause("sync", "", nil)

	req := httptest.NewRequest(http.MethodDelete, "/pause?job=sync", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if store.IsPaused("sync") {
		t.Error("expected job to be resumed")
	}
}

func TestPauseHandler_Resume_MissingJob(t *testing.T) {
	h, _ := buildPauseHandler()
	req := httptest.NewRequest(http.MethodDelete, "/pause", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPauseHandler_Resume_UnknownJob(t *testing.T) {
	h, _ := buildPauseHandler()
	req := httptest.NewRequest(http.MethodDelete, "/pause?job=ghost", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPauseHandler_MethodNotAllowed(t *testing.T) {
	h, _ := buildPauseHandler()
	req := httptest.NewRequest(http.MethodPut, "/pause", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestPauseHandler_InvalidBody(t *testing.T) {
	h, _ := buildPauseHandler()
	req := httptest.NewRequest(http.MethodPost, "/pause", bytes.NewBufferString(`{}`))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
