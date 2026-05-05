package config_test

import (
	"os"
	"testing"

	"github.com/yourorg/cronwatcher/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatcher-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
log_level: info
check_interval: 30s
jobs:
  - name: backup
    schedule: "0 2 * * *"
    command: /usr/local/bin/backup.sh
    timeout: 10m
alerting:
  email: ops@example.com
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", cfg.Jobs[0].Name)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `log_level: debug\njobs: []\n`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty jobs list")
	}
}

func TestLoad_MissingJobName(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - schedule: "* * * * *"
    command: /bin/true
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing job name")
	}
}
