# audit

Go/Goose microservice for audit logging and activity aggregation.

## Overview

`audit` records every create/update/delete event emitted by other services and exposes a paginated activity feed. A background `SummaryBatch` runs hourly (and on `apps.cron.heartbeat.hourly`) to aggregate log metrics.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Health check |
| GET | `/activities` | Paginated list of described audit log entries |
| GET/POST/PATCH/DELETE | `/log` | CRUD for raw audit log records |

## Entities

- **Log** — userId, clientId, workspaceId, group, service, entityId, entityName, action, desc, meta, status

## Event Subscriptions

| Pattern | Handler |
|---------|---------|
| `*.*.*.after-insert` | creates audit log |
| `*.*.*.after-update` | creates audit log |
| `*.*.*.after-delete` | creates audit log |
| `apps.cron.heartbeat.hourly` | runs SummaryBatch |
| `apps.cron.heartbeat.weekly/monthly/yearly` | creates audit log |

## Background Jobs

**SummaryBatch** — aggregates log counts by service, entity, and action over the previous hour. Enabled when `AUDIT_BATCH_TOGGLE=on` (or unset).

| Variable | Default | Description |
|----------|---------|-------------|
| `AUDIT_BATCH_TOGGLE` | `on` | Set to anything other than `on` to disable the batch |

## Module

```go
import "github.com/thescaffold/gox-apps-audit/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
