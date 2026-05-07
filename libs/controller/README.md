# controller

Go/Goose microservice for API gateway request routing and route management.

## Overview

`controller` is the API gateway layer. It stores Route definitions and incoming Request logs, and provides endpoints to inspect the current request context and resolve routes.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/requests` | Request log CRUD |
| GET/POST/PATCH/DELETE | `/routes` | Route CRUD |
| GET | `/now` | Return current timestamp / request context |
| GET | `/route` | Resolve a route by method + path |

## Entities

- **Request** — incoming request record (method, path, headers, status, latency)
- **Route** — registered route definition (method, path, service, action)

## Module

```go
import "github.com/thescaffold/gox-apps-controller/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
