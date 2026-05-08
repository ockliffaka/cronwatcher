package audit

import (
	"sync"
	"time"
)

// EventKind describes the type of audit event.
type EventKind string

const (
	EventJobStarted  EventKind = "job_started"
	EventJobFinished EventKind = "job_finished"
	EventAlertSent   EventKind = "alert_sent"
	EventConfigLoad  EventKind = "config_loaded"
)

// Event represents a single audit log entry.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      EventKind `json:"kind"`
	JobName   string    `json:"job_name,omitempty"`
	Message   string    `json:"message"`
}

// Log is an in-memory, bounded audit log.
type Log struct {
	mu       sync.RWMutex
	events   []Event
	maxItems int
}

// New creates a new audit Log that retains at most maxItems events.
func New(maxItems int) *Log {
	if maxItems <= 0 {
		maxItems = 200
	}
	return &Log{maxItems: maxItems}
}

// Record appends an event to the log, evicting the oldest entry when full.
func (l *Log) Record(kind EventKind, jobName, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		JobName:   jobName,
		Message:   message,
	}
	if len(l.events) >= l.maxItems {
		l.events = l.events[1:]
	}
	l.events = append(l.events, e)
}

// All returns a copy of all stored events, oldest first.
func (l *Log) All() []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// Filter returns events matching the given kind. Pass an empty string to return all.
func (l *Log) Filter(kind EventKind) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Event
	for _, e := range l.events {
		if kind == "" || e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}
