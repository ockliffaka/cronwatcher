package alert_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatcher/internal/alert"
	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/notify"
)

func makeStore(t *testing.T) *history.Store {
	t.Helper()
	return history.New(10)
}

func TestEvaluate_SendsAlert_OnFailure(t *testing.T) {
	var called int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := makeStore(t)
	store.Record("backup", history.Entry{ExitCode: 1, FinishedAt: time.Now()})

	mgr := alert.NewManager(store, notify.NewWebhookNotifier(srv.URL), time.Minute)
	if err := mgr.Evaluate("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", called)
	}
}

func TestEvaluate_NoAlert_OnSuccess(t *testing.T) {
	var called int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := makeStore(t)
	store.Record("backup", history.Entry{ExitCode: 0, FinishedAt: time.Now()})

	mgr := alert.NewManager(store, notify.NewWebhookNotifier(srv.URL), time.Minute)
	if err := mgr.Evaluate("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("expected no webhook call on success, got %d", called)
	}
}

func TestEvaluate_NoAlert_UnknownJob(t *testing.T) {
	store := makeStore(t)
	mgr := alert.NewManager(store, notify.NewWebhookNotifier("http://localhost"), time.Minute)
	if err := mgr.Evaluate("nonexistent"); err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestEvaluate_RateLimits_DuplicateAlerts(t *testing.T) {
	var called int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := makeStore(t)
	store.Record("backup", history.Entry{ExitCode: 2, FinishedAt: time.Now()})

	mgr := alert.NewManager(store, notify.NewWebhookNotifier(srv.URL), time.Minute)
	_ = mgr.Evaluate("backup")
	_ = mgr.Evaluate("backup")
	_ = mgr.Evaluate("backup")

	if n := atomic.LoadInt32(&called); n != 1 {
		t.Fatalf("expected 1 webhook call due to rate limiting, got %d", n)
	}
}
