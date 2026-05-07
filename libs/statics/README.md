# statics

Go/Goose microservice for static lookup lists (country codes, currencies, enums, etc.).

## Overview

`statics` is a read-mostly service that seeds and serves structured static data lists. Other services query it to resolve display names, codes, and allowed values without hardcoding them.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/list` | List CRUD |
| GET | `/filter/:key` | Filter list entries by key |
| GET | `/code/:code` | Lookup by code |
| GET | `/value/:value` | Lookup by value |

## Entities

- **List** — key, code, value, label, meta, status

## Module

```go
import "github.com/thescaffold/gox-apps-statics/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
