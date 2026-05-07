# cron

Go/Goose microservice for distributed cron job scheduling.

## Overview

`cron` is the scheduler service. Other services register recurring jobs here; the cron service selects jobs that are due, dispatches them, and records execution logs. It acts as the heartbeat source for `apps.cron.heartbeat.*` events consumed by other services.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/jobs` | Job CRUD |
| GET/POST/PATCH/DELETE | `/logs` | Log CRUD |
| POST | `/` | Register a new cron job |
| GET | `/select` | Select jobs that are due to run |
| PATCH | `/log` | Update a job execution log |

## Entities

- **Job** — name, schedule (cron expression), service, action, status
- **Log** — jobId, status, output, duration

## Module

```go
import "github.com/thescaffold/gox-apps-cron/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
