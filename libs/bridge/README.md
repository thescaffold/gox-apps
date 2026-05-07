# bridge

Go/Goose microservice for license, subscription plan, preference, and webhook management.

## Overview

`bridge` is the billing and subscription gateway. It manages License records (which plans a client has activated), LicenseTypes (pricing tiers), Preferences (per-client config), Webhooks (outbound event hooks), and WebhookLogs.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/licenses` | License CRUD |
| GET/POST/PATCH/DELETE | `/licensetypes` | LicenseType CRUD |
| GET/POST/PATCH/DELETE | `/plantypes` | PlanType CRUD |
| GET/POST/PATCH/DELETE | `/preferences` | Preference CRUD |
| GET/POST/PATCH/DELETE | `/webhooks` | Webhook CRUD |
| GET/POST/PATCH/DELETE | `/webhooklogs` | WebhookLog CRUD |

## Entities

- **License** — activation record linking a client to a LicenseType
- **LicenseType** — pricing tier with daily/weekly/monthly/yearly rates and currency
- **PlanType** — plan category
- **Preference** — key/value configuration per client
- **Webhook** — outbound HTTP hook configuration
- **WebhookLog** — delivery attempt record

## Module

```go
import "github.com/thescaffold/gox-apps-bridge/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
