package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cronwatcher/internal/alert"
	"github.com/cronwatcher/internal/api"
	"github.com/cronwatcher/internal/history"
)

type fakeNotifier struct{ err error }

func (f *fakeNotifier) Send(_ string) error { return f.err }

func buildAlertHandler(exitCode int, notifyErr error) http.HandlerFunc {
	s := history.New(10)
	s.Record(history.Entry{
		JobName:    "myjob",
		ExitCode:   exitCode,
		StartedAt:  time.Now().Add(-2 * time.Second),
		FinishedAt: time.Now(),
	})
	mgr := alert.NewManager(&fakeNotifier{err: notifyErr}, s)
	return api.NewAlertHandler(mgr)
}

func TestAlertHandler_Success(t *testing.T) {
	h := buildAlertHandler(0, nil)
	body := `{"job_name":"myjob"}`
	req := httptest.NewRequest(http.MethodPost, "/alert", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestAlertHandler_MethodNotAllowed(t *testing.T) {
	h := buildAlertHandler(0, nil)
	req := httptest.NewRequest(http.MethodGet, "/alert", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestAlertHandler_MissingJobName(t *testing.T) {
	h := buildAlertHandler(0, nil)
	req := httptest.NewRequest(http.MethodPost, "/alert", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAlertHandler_NotifierError(t *testing.T) {
	h := buildAlertHandler(1, errors.New("webhook down"))
	body := `{"job_name":"myjob"}`
	req := httptest.NewRequest(http.MethodPost, "/alert", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
