# queue

Go/Goose microservice for distributed job queuing.

## Overview

`queue` implements a database-backed FIFO job queue. Services push jobs and pop them in priority order. Job execution is tracked via Log records. Queue metadata (name, priority, concurrency) is managed through the Queue resource.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/jobs` | Job CRUD |
| GET/POST/PATCH/DELETE | `/logs` | Log CRUD |
| GET/POST/PATCH/DELETE | `/queues` | Queue CRUD |
| POST | `/` | Push a new job onto a queue |
| GET | `/pop` | Pop the next due job from a queue |
| PATCH | `/log` | Update a job execution log |

## Entities

- **Job** — queueId, name, payload, priority, status, scheduledAt
- **Log** — jobId, status, output, duration
- **Queue** — name, concurrency, status

## Module

```go
import "github.com/thescaffold/gox-apps-queue/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
