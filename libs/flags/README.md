# flags

Go/Goose microservice for feature flag management.

## Overview

`flags` stores feature flag definitions, environments, and flag logs. It integrates with the `gox-packages/flags` package to register, log, check status, and apply rate limits on flags.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/environments` | Environment CRUD |
| GET/POST/PATCH/DELETE | `/environmenttypes` | EnvironmentType CRUD |
| GET/POST/PATCH/DELETE | `/flags` | Flag CRUD |
| GET/POST/PATCH/DELETE | `/flaglogs` | FlagLog CRUD |

## Entities

- **Environment** — name, type, status
- **EnvironmentType** — name, meta
- **Flag** — name, key, environment, value, status
- **FlagLog** — flagId, action, meta

## Module

```go
import "github.com/thescaffold/gox-apps-flags/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
