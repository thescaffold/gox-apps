# cache

Go/Goose microservice for key-value cache storage.

## Overview

`cache` provides a database-backed key-value store with named lists. It exposes both CRUD endpoints for managing list records and an operational API (get/set/del/flush) that mirrors a simple cache interface.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Health check |
| GET/POST/PATCH/DELETE | `/list` | CRUD for List records |
| GET | `/get` | Retrieve a value by key |
| POST | `/set` | Store a key-value pair |
| DELETE | `/del` | Delete a key |
| DELETE | `/flush` | Clear all keys in a list |

## Entities

- **List** — name, key, value, status

## Module

```go
import "github.com/thescaffold/gox-apps-cache/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
