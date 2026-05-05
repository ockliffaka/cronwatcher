package notify

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AlertPayload holds the information about a cron job failure.
type AlertPayload struct {
	JobName   string
	ExitCode  int
	Output    string
	Timestamp time.Time
}

// Notifier defines the interface for sending alerts.
type Notifier interface {
	Send(payload AlertPayload) error
}

// WebhookNotifier sends alerts to an HTTP webhook endpoint.
type WebhookNotifier struct {
	URL        string
	HTTPClient *http.Client
}

// NewWebhookNotifier creates a new WebhookNotifier with the given URL.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL: url,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send posts a JSON alert payload to the configured webhook URL.
func (w *WebhookNotifier) Send(payload AlertPayload) error {
	body := fmt.Sprintf(
		`{"job":%q,"exit_code":%d,"output":%q,"timestamp":%q}`,
		payload.JobName,
		payload.ExitCode,
		payload.Output,
		payload.Timestamp.Format(time.RFC3339),
	)

	resp, err := w.HTTPClient.Post(w.URL, "application/json", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
