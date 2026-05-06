package history

import (
	"log"
	"time"
)

// Cleaner periodically removes history entries older than a given retention period.
type Cleaner struct {
	store     *Store
	retention time.Duration
	interval  time.Duration
	stop      chan struct{}
}

// NewCleaner creates a Cleaner that will purge entries older than retention,
// running the sweep every interval.
func NewCleaner(store *Store, retention, interval time.Duration) *Cleaner {
	return &Cleaner{
		store:     store,
		retention: retention,
		interval:  interval,
		stop:      make(chan struct{}),
	}
}

// Start launches the background sweep goroutine.
func (c *Cleaner) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				removed := c.sweep()
				if removed > 0 {
					log.Printf("history cleaner: removed %d stale entries", removed)
				}
			case <-c.stop:
				return
			}
		}
	}()
}

// Stop halts the background goroutine.
func (c *Cleaner) Stop() {
	close(c.stop)
}

// sweep removes all entries older than the retention window and returns the
// count of removed entries.
func (c *Cleaner) sweep() int {
	cutoff := time.Now().Add(-c.retention)
	c.store.mu.Lock()
	defer c.store.mu.Unlock()

	removed := 0
	for job, entries := range c.store.entries {
		var kept []Entry
		for _, e := range entries {
			if e.StartedAt.After(cutoff) {
				kept = append(kept, e)
			} else {
				removed++
			}
		}
		c.store.entries[job] = kept
	}
	return removed
}
