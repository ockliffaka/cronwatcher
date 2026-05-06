package history

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func buildExporter(t *testing.T) (*Exporter, *Store) {
	t.Helper()
	s := New(10)
	now := time.Now()
	s.Record("backup", Entry{Success: true, StartedAt: now, Duration: 2 * time.Second, Output: "ok"})
	s.Record("backup", Entry{Success: false, StartedAt: now.Add(-time.Hour), Duration: 500 * time.Millisecond, Output: "err"})
	s.Record("cleanup", Entry{Success: true, StartedAt: now, Duration: time.Second, Output: ""})
	return NewExporter(s), s
}

func TestExporter_JSON_ContainsAllEntries(t *testing.T) {
	ex, _ := buildExporter(t)
	var buf bytes.Buffer
	if err := ex.Export(&buf, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []ExportEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestExporter_CSV_HasHeader(t *testing.T) {
	ex, _ := buildExporter(t)
	var buf bytes.Buffer
	if err := ex.Export(&buf, FormatCSV); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if lines[0] != "job_name,success,started_at,duration_ms,output" {
		t.Errorf("unexpected CSV header: %s", lines[0])
	}
	// header + 3 data rows
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestExporter_CSV_ContainsJobName(t *testing.T) {
	ex, _ := buildExporter(t)
	var buf bytes.Buffer
	_ = ex.Export(&buf, FormatCSV)

	if !strings.Contains(buf.String(), "backup") {
		t.Error("expected CSV to contain job name 'backup'")
	}
}

func TestExporter_UnsupportedFormat(t *testing.T) {
	ex, _ := buildExporter(t)
	var buf bytes.Buffer
	err := ex.Export(&buf, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExporter_EmptyStore(t *testing.T) {
	s := New(10)
	ex := NewExporter(s)
	var buf bytes.Buffer
	if err := ex.Export(&buf, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var entries []ExportEntry
	_ = json.Unmarshal(buf.Bytes(), &entries)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty store, got %d", len(entries))
	}
}
