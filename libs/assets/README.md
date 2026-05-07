# assets

Go/Goose microservice for static file and dynamic SVG asset management.

## Overview

`assets` stores uploaded files and generates on-demand SVG images (pixel, gradient, solid variants). It is the file-hosting backbone used by other services that need to serve images or attachments.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Health check |
| GET | `/dynamic` | Generate or fetch a dynamic SVG by ID or parameters |
| GET/POST/PATCH/DELETE | `/file` | CRUD for stored File records |

## Entities

- **File** — type, bucket, name, tags, raw SVG content, status

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | Postgres/MySQL connection string |
| `JWT_SECRET` | `changeme` | JWT signing secret |

## Module

```go
import "github.com/thescaffold/gox-apps-assets/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
