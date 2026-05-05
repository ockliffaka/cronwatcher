package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level cronwatcher configuration.
type Config struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	LogLevel      string        `yaml:"log_level"`
	Jobs          []JobConfig   `yaml:"jobs"`
	Alerting      AlertConfig   `yaml:"alerting"`
}

// JobConfig describes a single monitored cron job.
type JobConfig struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Timeout  time.Duration `yaml:"timeout"`
	Command  string        `yaml:"command"`
}

// AlertConfig holds alerting destination settings.
type AlertConfig struct {
	Email   string `yaml:"email"`
	SlackURL string `yaml:"slack_url"`
}

// Load reads and parses the YAML configuration file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate performs basic semantic checks on the loaded configuration.
func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("at least one job must be defined")
	}
	for i, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
		if job.Command == "" {
			return fmt.Errorf("job %q: command is required", job.Name)
		}
	}
	return nil
}
