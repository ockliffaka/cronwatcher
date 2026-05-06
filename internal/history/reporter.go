package history

import (
	"fmt"
	"strings"
	"time"
)

// Summary aggregates statistics for a single job.
type Summary struct {
	JobName      string
	Total        int
	Successes    int
	Failures     int
	LastStatus   Status
	LastRun      time.Time
	AvgDuration  time.Duration
}

// Summarise computes a Summary for the given job from its stored entries.
// Returns false if no entries exist for the given job name.
func (s *Store) Summarise(jobName string) (Summary, bool) {
	entries := s.Get(jobName)
	if len(entries) == 0 {
		return Summary{}, false
	}

	var totalDur time.Duration
	var successes, failures int

	for _, e := range entries {
		totalDur += e.Duration
		switch e.Status {
		case StatusSuccess:
			successes++
		case StatusFailure:
			failures++
		}
	}

	last := entries[len(entries)-1]
	return Summary{
		JobName:     jobName,
		Total:       len(entries),
		Successes:   successes,
		Failures:    failures,
		LastStatus:  last.Status,
		LastRun:     last.StartedAt,
		AvgDuration: totalDur / time.Duration(len(entries)),
	}, true
}

// SuccessRate returns the percentage of successful runs as a value between
// 0.0 and 100.0. Returns 0.0 if there are no runs recorded.
func (sum Summary) SuccessRate() float64 {
	if sum.Total == 0 {
		return 0.0
	}
	return float64(sum.Successes) / float64(sum.Total) * 100.0
}

// Format returns a human-readable string for a Summary.
func (sum Summary) Format() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Job: %s\n", sum.JobName)
	fmt.Fprintf(&sb, "  Total runs : %d\n", sum.Total)
	fmt.Fprintf(&sb, "  Successes  : %d\n", sum.Successes)
	fmt.Fprintf(&sb, "  Failures   : %d\n", sum.Failures)
	fmt.Fprintf(&sb, "  Success rate: %.1f%%\n", sum.SuccessRate())
	fmt.Fprintf(&sb, "  Last status: %s\n", sum.LastStatus)
	fmt.Fprintf(&sb, "  Last run   : %s\n", sum.LastRun.Format(time.RFC3339))
	fmt.Fprintf(&sb, "  Avg duration: %s\n", sum.AvgDuration)
	return sb.String()
}
