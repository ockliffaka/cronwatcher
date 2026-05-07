package alert

import (
	"fmt"
	"time"

	"github.com/cronwatcher/internal/history"
	"github.com/cronwatcher/internal/notify"
)

// Severity represents the alert level.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Alert holds the details of a triggered alert.
type Alert struct {
	JobName   string
	Severity  Severity
	Message   string
	OccurredAt time.Time
}

// Manager evaluates job history and fires alerts via a Notifier.
type Manager struct {
	notifier notify.Notifier
	store    *history.Store
}

// NewManager creates a new alert Manager.
func NewManager(n notify.Notifier, s *history.Store) *Manager {
	return &Manager{notifier: n, store: s}
}

// Evaluate checks the latest run for the given job and sends an alert if it failed.
func (m *Manager) Evaluate(jobName string) error {
	entry, ok := m.store.Last(jobName)
	if !ok {
		return nil
	}

	if entry.ExitCode == 0 {
		return nil
	}

	a := Alert{
		JobName:    jobName,
		Severity:   SeverityError,
		Message:    fmt.Sprintf("job %q failed with exit code %d at %s", jobName, entry.ExitCode, entry.FinishedAt.Format(time.RFC3339)),
		OccurredAt: entry.FinishedAt,
	}

	return m.notifier.Send(a.Message)
}
