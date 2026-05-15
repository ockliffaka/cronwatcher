package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/cronwatcher/internal/api"
	"github.com/example/cronwatcher/internal/concurrency"
)

func buildConcurrencyHandler() (http.Handler, *concurrency.Store) {
	s := concurrency.New()
	return api.NewConcurrencyHandler(s), s
}

func TestConcurrencyHandler_SetAndGet(t *testing.T) {
	h, _ := buildConcurrencyHandler()

	body := `{"job":"backup","max":3}`
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/concurrency", strings.NewReader(body)))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/concurrency?job=backup", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entry concurrency.Entry
	if err := json.NewDecoder(rr.Body).Decode(&entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.Max != 3 {
		t.Fatalf("expected max=3, got %d", entry.Max)
	}
}

func TestConcurrencyHandler_GetAll(t *testing.T) {
	h, s := buildConcurrencyHandler()
	_ = s.SetMax("job-x", 2)
	_ = s.SetMax("job-y", 4)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/concurrency", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var all map[string]concurrency.Entry
	if err := json.NewDecoder(rr.Body).Decode(&all); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestConcurrencyHandler_GetUnknownJob(t *testing.T) {
	h, _ := buildConcurrencyHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/concurrency?job=ghost", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestConcurrencyHandler_Release(t *testing.T) {
	h, s := buildConcurrencyHandler()
	_ = s.SetMax("job-r", 1)
	_ = s.Acquire("job-r")

	body := bytes.NewBufferString(`{"job":"job-r"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/concurrency/release", body)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	e, _ := s.Get("job-r")
	if e.Active != 0 {
		t.Fatalf("expected active=0 after release, got %d", e.Active)
	}
}

func TestConcurrencyHandler_MethodNotAllowed(t *testing.T) {
	h, _ := buildConcurrencyHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/api/concurrency", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestConcurrencyHandler_MissingJob_Returns400(t *testing.T) {
	h, _ := buildConcurrencyHandler()
	body := bytes.NewBufferString(`{"max":2}`)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/concurrency", body))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
