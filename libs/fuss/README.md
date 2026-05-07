# fuss

Go/Goose microservice for search history, tokens, and lookup.

## Overview

`fuss` provides a user search history and token management service. It records what users searched, exposes recent search history, and manages searchable tokens with lookup and full-text-style search endpoints.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/histories` | History CRUD |
| GET/POST/PATCH/DELETE | `/tokens` | Token CRUD |
| GET/POST/PATCH/DELETE | `/token-logs` | TokenLog CRUD |
| GET | `/recent` | Recent search history for the current user |
| GET | `/search` | Full-text token search |
| GET | `/lookup` | Exact token lookup |

## Entities

- **History** — userId, query, status
- **Token** — value, type, meta, status
- **TokenLog** — tokenId, userId, action

## Module

```go
import "github.com/thescaffold/gox-apps-fuss/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
