package history

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportFormat defines the supported export formats.
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// ExportEntry is a flattened representation of a history entry for export.
type ExportEntry struct {
	JobName   string        `json:"job_name"`
	Success   bool          `json:"success"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ms"`
	Output    string        `json:"output,omitempty"`
}

// Exporter writes job history to an output stream in the requested format.
type Exporter struct {
	store *Store
}

// NewExporter creates a new Exporter backed by the given Store.
func NewExporter(s *Store) *Exporter {
	return &Exporter{store: s}
}

// Export writes all history entries for every known job to w in the given format.
func (e *Exporter) Export(w io.Writer, format ExportFormat) error {
	entries := e.collect()

	switch format {
	case FormatJSON:
		return e.writeJSON(w, entries)
	case FormatCSV:
		return e.writeCSV(w, entries)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

func (e *Exporter) collect() []ExportEntry {
	e.store.mu.RLock()
	defer e.store.mu.RUnlock()

	var out []ExportEntry
	for name, entries := range e.store.records {
		for _, en := range entries {
			out = append(out, ExportEntry{
				JobName:   name,
				Success:   en.Success,
				StartedAt: en.StartedAt,
				Duration:  en.Duration,
				Output:    en.Output,
			})
		}
	}
	return out
}

func (e *Exporter) writeJSON(w io.Writer, entries []ExportEntry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func (e *Exporter) writeCSV(w io.Writer, entries []ExportEntry) error {
	_, err := fmt.Fprintln(w, "job_name,success,started_at,duration_ms,output")
	if err != nil {
		return err
	}
	for _, en := range entries {
		_, err = fmt.Fprintf(w, "%s,%t,%s,%d,%s\n",
			en.JobName,
			en.Success,
			en.StartedAt.Format(time.RFC3339),
			en.Duration.Milliseconds(),
			en.Output,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
