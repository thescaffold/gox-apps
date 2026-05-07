# common

Go/Goose microservice for shared reference data: IPs, currency rates, tags, and projects.

## Overview

`common` stores cross-service reference data. It provides geo-IP lookup (backed by ip-api.com with DB caching), currency exchange rates, and tag/project classification records used by other services.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/ips` | IP record CRUD |
| GET/POST/PATCH/DELETE | `/projects` | Project CRUD |
| GET/POST/PATCH/DELETE | `/project-types` | ProjectType CRUD |
| GET/POST/PATCH/DELETE | `/rates` | Rate CRUD |
| GET/POST/PATCH/DELETE | `/rate-logs` | RateLog CRUD |
| GET/POST/PATCH/DELETE | `/tags` | Tag CRUD |
| GET/POST/PATCH/DELETE | `/tag-types` | TagType CRUD |
| GET | `/currency/:currency` | Lookup exchange rate by currency code |
| GET | `/location?ip=<ip>` | Geo-IP lookup (cached in DB) |

## Entities

Ip, Project, ProjectType, Rate, RateLog, Tag, TagType

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GEO_API_URL` | `http://ip-api.com/json` | Base URL for geo-IP lookups |

## Module

```go
import "github.com/thescaffold/gox-apps-common/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
