package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatcher/internal/tag"
)

func buildTagHandler() (http.HandlerFunc, *tag.Store) {
	s := tag.New()
	return NewTagHandler(s), s
}

func TestTagHandler_SetAndGet(t *testing.T) {
	h, _ := buildTagHandler()

	body := `{"job":"backup","tags":["critical","db"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/tags?job=backup", nil)
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec2.Body).Decode(&resp)
	if resp["job"] != "backup" {
		t.Errorf("unexpected job in response: %v", resp)
	}
}

func TestTagHandler_GetUnknownJob(t *testing.T) {
	h, _ := buildTagHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/tags?job=ghost", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTagHandler_GetAll(t *testing.T) {
	h, s := buildTagHandler()
	s.Set("job1", []string{"a"})
	s.Set("job2", []string{"b"})

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string][]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 jobs in response, got %d", len(resp))
	}
}

func TestTagHandler_Delete(t *testing.T) {
	h, s := buildTagHandler()
	s.Set("job", []string{"x"})

	req := httptest.NewRequest(http.MethodDelete, "/api/tags?job=job", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, err := s.Get("job")
	if err == nil {
		t.Error("expected job to be deleted from store")
	}
}

func TestTagHandler_MethodNotAllowed(t *testing.T) {
	h, _ := buildTagHandler()
	req := httptest.NewRequest(http.MethodPatch, "/api/tags", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestTagHandler_PostMissingJob(t *testing.T) {
	h, _ := buildTagHandler()
	body := `{"tags":["x"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
