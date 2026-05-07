package alert

import (
	"fmt"
	"log"
	"time"

	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/notify"
	"github.com/cronwatcher/internal/ratelimit"
)

// Manager evaluates job history and dispatches alerts via a Notifier.
type Manager struct {
	store    *history.Store
	notifier *notify.WebhookNotifier
	limiter  *ratelimit.Limiter
}

// NewManager creates an alert Manager. cooldown controls how frequently
// repeated failure alerts are suppressed for the same job.
func NewManager(store *history.Store, notifier *notify.WebhookNotifier, cooldown time.Duration) *Manager {
	return &Manager{
		store:    store,
		notifier: notifier,
		limiter:  ratelimit.New(cooldown),
	}
}

// Evaluate checks the most recent run of jobName and sends an alert if
// it failed, subject to the rate limiter.
func (m *Manager) Evaluate(jobName string) error {
	entry, ok := m.store.Last(jobName)
	if !ok {
		return fmt.Errorf("alert: no history for job %q", jobName)
	}

	if entry.ExitCode == 0 {
		m.limiter.Reset(jobName)
		return nil
	}

	if !m.limiter.Allow(jobName) {
		log.Printf("alert: suppressed duplicate alert for job %q (rate limited)", jobName)
		return nil
	}

	msg := fmt.Sprintf("[cronwatcher] job %q failed with exit code %d at %s",
		jobName, entry.ExitCode, entry.FinishedAt.Format(time.RFC3339))

	if err := m.notifier.Send(msg); err != nil {
		return fmt.Errorf("alert: failed to send notification: %w", err)
	}
	return nil
}
