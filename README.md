# gox-apps

Go/Goose microservices — the Go rewrite of the `ntx-apps` NestJS monorepo.

## Overview

`gox-apps` is a Go workspace containing 19 independent microservices built with the [Goose](https://github.com/awesome-goose/goose) framework. Each service lives in `libs/<name>/` as a self-contained Go module and can be deployed standalone or composed into a larger platform.

## Services

| Service | Module | Purpose |
|---------|--------|---------|
| [assets](libs/assets/README.md) | `gox-apps-assets` | Static file storage and dynamic SVG generation |
| [audit](libs/audit/README.md) | `gox-apps-audit` | Audit logging and hourly activity aggregation |
| [blobs](libs/blobs/README.md) | `gox-apps-blobs` | Binary file and multi-page document storage |
| [bridge](libs/bridge/README.md) | `gox-apps-bridge` | License, subscription plan, and webhook management |
| [cache](libs/cache/README.md) | `gox-apps-cache` | Database-backed key-value cache |
| [capital](libs/capital/README.md) | `gox-apps-capital` | Payments, accounts, wallets, and transactions |
| [common](libs/common/README.md) | `gox-apps-common` | Shared reference data (IPs, rates, tags, projects) |
| [controller](libs/controller/README.md) | `gox-apps-controller` | API gateway request routing |
| [cron](libs/cron/README.md) | `gox-apps-cron` | Distributed cron job scheduling |
| [figs](libs/figs/README.md) | `gox-apps-figs` | File conversion (CSV/Excel/PDF), validation, and storage |
| [flags](libs/flags/README.md) | `gox-apps-flags` | Feature flag management |
| [forms](libs/forms/README.md) | `gox-apps-forms` | Dynamic form definitions and submissions |
| [fuss](libs/fuss/README.md) | `gox-apps-fuss` | Search history and token lookup |
| [health](libs/health/README.md) | `gox-apps-health` | Service health monitoring with auto-summary |
| [identity](libs/identity/README.md) | `gox-apps-identity` | Auth (JWT/Basic/Token/HMAC), users, roles, permissions |
| [notification](libs/notification/README.md) | `gox-apps-notification` | Multi-channel notification delivery (email, web) |
| [polylog](libs/polylog/README.md) | `gox-apps-polylog` | Event streaming pipeline and webhook sink dispatch |
| [queue](libs/queue/README.md) | `gox-apps-queue` | Database-backed distributed job queue |
| [statics](libs/statics/README.md) | `gox-apps-statics` | Static lookup lists (codes, values, labels) |

## Repository Structure

```
gox-apps/
├── go.work               # Go workspace linking all service modules
├── go.work.sum
├── libs/
│   ├── <service>/
│   │   ├── go.mod        # Independent Go module
│   │   ├── app/          # Goose module, controller, service, routes, DTOs
│   │   ├── migrations/   # SQL schema migrations
│   │   ├── pkg/          # Optional service-specific sub-packages
│   │   └── tests/        # Integration tests
│   └── ...
└── Makefile
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL (or SQLite for local dev)

### Build all services

```bash
go build ./...
```

### Test all services

```bash
go test ./...
```

### Run a service

Each service exposes a goose `AppModule`. Wire it into a goose `Application`:

```go
package main

import (
    "github.com/awesome-goose/goose"
    assets "github.com/thescaffold/gox-apps-assets/app"
)

func main() {
    app := goose.New(&assets.AppModule{})
    app.Listen(":3000")
}
```

## Environment Variables (common across services)

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | Postgres connection string |
| `JWT_SECRET` | `changeme` | JWT signing secret |
| `HMAC_KEY` | `changeme-hmac` | HMAC signing key for internal calls |
| `PORT` | `3000` | HTTP listen port |

## Architecture

Services communicate via:
- **In-process event bus** — `core/events.Bus` for same-process pub/sub
- **HTTP** — `core/http.Client` for cross-service calls (HMAC-authenticated internally)
- **Database** — each service owns its own tables; no cross-service DB joins

See [VERIFY.md](../VERIFY.md) for a full feature-parity comparison with the original NestJS implementation.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
