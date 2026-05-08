package audit

import (
	"testing"
)

func TestRecord_StoresEvent(t *testing.T) {
	l := New(10)
	l.Record(EventJobStarted, "backup", "started")
	events := l.All()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Kind != EventJobStarted {
		t.Errorf("expected kind %q, got %q", EventJobStarted, events[0].Kind)
	}
	if events[0].JobName != "backup" {
		t.Errorf("unexpected job name: %s", events[0].JobName)
	}
}

func TestRecord_Eviction(t *testing.T) {
	l := New(3)
	for i := 0; i < 5; i++ {
		l.Record(EventAlertSent, "job", "msg")
	}
	if len(l.All()) != 3 {
		t.Errorf("expected 3 events after eviction, got %d", len(l.All()))
	}
}

func TestFilter_ByKind(t *testing.T) {
	l := New(20)
	l.Record(EventJobStarted, "a", "start")
	l.Record(EventAlertSent, "a", "alert")
	l.Record(EventJobFinished, "a", "done")

	results := l.Filter(EventAlertSent)
	if len(results) != 1 {
		t.Fatalf("expected 1 alert event, got %d", len(results))
	}
	if results[0].Kind != EventAlertSent {
		t.Errorf("unexpected kind: %s", results[0].Kind)
	}
}

func TestFilter_EmptyKind_ReturnsAll(t *testing.T) {
	l := New(20)
	l.Record(EventJobStarted, "x", "s")
	l.Record(EventConfigLoad, "", "loaded")

	results := l.Filter("")
	if len(results) != 2 {
		t.Errorf("expected 2 events, got %d", len(results))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := New(10)
	l.Record(EventJobFinished, "b", "ok")
	copy1 := l.All()
	copy1[0].Message = "mutated"
	copy2 := l.All()
	if copy2[0].Message == "mutated" {
		t.Error("All() should return an independent copy")
	}
}
