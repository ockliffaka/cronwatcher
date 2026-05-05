# cronwatcher

Lightweight daemon that monitors cron job execution and sends alerts on failures.

## Installation

```bash
go install github.com/cronwatcher/cronwatcher@latest
```

Or build from source:

```bash
git clone https://github.com/cronwatcher/cronwatcher.git && cd cronwatcher && go build ./...
```

## Usage

Create a configuration file `config.yaml`:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    command: "/usr/local/bin/backup.sh"
    timeout: 300
    alert:
      email: "ops@example.com"

alerts:
  smtp:
    host: "smtp.example.com"
    port: 587
    from: "cronwatcher@example.com"
```

Start the daemon:

```bash
cronwatcher --config config.yaml
```

cronwatcher will track each job's exit code, duration, and last run time. If a job fails or exceeds its timeout, an alert is dispatched immediately.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.yaml` | Path to configuration file |
| `--log-level` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |
| `--dry-run` | `false` | Validate config without starting the daemon |

## License

MIT © cronwatcher contributors