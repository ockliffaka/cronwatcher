package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwatcher/internal/notify"
)

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := notify.NewWebhookNotifier(server.URL)
	payload := notify.AlertPayload{
		JobName:   "backup-job",
		ExitCode:  1,
		Output:    "disk full",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	if err := notifier.Send(payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["job"] != "backup-job" {
		t.Errorf("expected job=backup-job, got %v", received["job"])
	}
}

func TestWebhookNotifier_Send_Non2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := notify.NewWebhookNotifier(server.URL)
	err := notifier.Send(notify.AlertPayload{
		JobName:   "test-job",
		ExitCode:  2,
		Timestamp: time.Now(),
	})

	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestWebhookNotifier_Send_InvalidURL(t *testing.T) {
	notifier := notify.NewWebhookNotifier("http://127.0.0.1:0/no-server")
	err := notifier.Send(notify.AlertPayload{
		JobName:   "test-job",
		ExitCode:  1,
		Timestamp: time.Now(),
	})

	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}
