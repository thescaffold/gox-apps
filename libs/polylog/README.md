# polylog

Go/Goose microservice for event streaming, pipeline routing, and sink dispatch.

## Overview

`polylog` is the event streaming backbone. It ingests events from any source, queues them for async processing, routes them through channel configurations to configured sinks (webhooks, email, SMS), and records EventLogs for observability.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/channels` | Channel CRUD |
| GET/POST/PATCH/DELETE | `/events` | Event CRUD |
| GET/POST/PATCH/DELETE | `/eventlogs` | EventLog CRUD |
| GET/POST/PATCH/DELETE | `/sinks` | Sink CRUD |
| GET/POST/PATCH/DELETE | `/sinktypes` | SinkType CRUD |
| GET/POST/PATCH/DELETE | `/sources` | Source CRUD |
| GET/POST/PATCH/DELETE | `/sourcetypes` | SourceType CRUD |
| POST | `/ingest` | Ingest an event and trigger async pipeline processing |

## Entities

Channel, Event, EventLog, Sink, SinkType, Source, SourceType

## Pipeline

`pkg/pipeline.PipelineService`:

1. **Ingest** — saves the Event record, spawns a goroutine for async processing
2. **Process** — loads all Channels matching the event type, calls Discharge on each
3. **Discharge** — routes to the appropriate Sink implementation
4. **WebhookSink** — HTTP POST to the sink's configured URL with the event payload

## Sink Types

| Type | Status |
|------|--------|
| `webhook` | Implemented |
| `at-sms`, `mailchimp-email`, `mailgun-email`, `sendgrid-email`, `form` | Stub (extensible) |

## Module

```go
import "github.com/thescaffold/gox-apps-polylog/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
