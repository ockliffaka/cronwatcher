package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier is the interface for sending alert messages.
type Notifier interface {
	Send(message string) error
}

// WebhookNotifier sends alerts to an HTTP webhook endpoint.
type WebhookNotifier struct {
	URL    string
	client *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a sensible default timeout.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts a JSON payload containing the message to the configured URL.
func (w *WebhookNotifier) Send(message string) error {
	payload, err := json.Marshal(map[string]string{"text": message})
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("notify: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notify: unexpected status %d", resp.StatusCode)
	}
	return nil
}
