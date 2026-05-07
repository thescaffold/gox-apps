# health

Go/Goose microservice for service health monitoring and summary computation.

## Overview

`health` tracks the availability of registered services by recording heartbeat logs and computing health summary scores. A background `SummaryTask` runs every 10 minutes to update each service's summary status automatically.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Health check |
| GET/POST/PATCH/DELETE | `/log` | Log CRUD |
| GET/POST/PATCH/DELETE | `/service` | Service CRUD |
| GET/POST/PATCH/DELETE | `/summary` | Summary CRUD |
| POST | `/check/:id` | Trigger an on-demand health check for service `id` |

## Entities

- **Log** — serviceId, status, latency, meta
- **Service** — name, url, type, status
- **Summary** — serviceId, type (Up/Troubled/PossibleTroubled/Down), measure, received

## Background Jobs

**SummaryTask** — runs every 10 minutes. For each registered service, counts log entries in the last interval and computes a health score. Updates the Summary record and the service's current type.

| Variable | Default | Description |
|----------|---------|-------------|
| `SUMMARY_TOGGLE` | `on` | Set to anything other than `on` to disable |
| `SUMMARY_INTERVAL_SECS` | `3600` | Lookback window in seconds |
| `SUMMARY_EXPECTED_EVENT_COUNT` | `12` | Expected log count for a healthy service |

## Module

```go
import "github.com/thescaffold/gox-apps-health/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
