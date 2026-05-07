# notification

Go/Goose microservice for multi-channel notification delivery.

## Overview

`notification` stores notification templates, rules, messages, and delivery logs. The `pkg/provider` sub-package implements actual delivery via Mailgun (email), Zoho Mail (email), in-app web messages, and stub SMS/mobile channels. Templates are rendered with Mustache before dispatch.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/logs` | NotificationLog CRUD |
| GET/POST/PATCH/DELETE | `/messages` | Message CRUD |
| GET/POST/PATCH/DELETE | `/rules` | Rule CRUD |
| GET/POST/PATCH/DELETE | `/templates` | Template CRUD |

## Entities

- **Log** — userId, channel, templateId, ruleId, status, meta
- **Message** — userId, title, body, type, status
- **Rule** — userId, channel, enabled, meta
- **Template** — key, group, subject, body (Mustache), channels, status

## Provider Package

`pkg/provider.ProviderService.Send(log)`:
1. Looks up the Template by `log.Key` + `log.Group`
2. Resolves notification rules for the user
3. Renders the template body with Mustache using `log.Data`
4. Dispatches to enabled channels: email (Mailgun/Zoho), web (Message record), SMS/mobile (stub)
5. Updates the Log record with delivery status

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MAILGUN_API_KEY` | — | Mailgun private API key |
| `MAILGUN_DOMAIN` | — | Mailgun sending domain |
| `MAILGUN_FROM` | — | Sender address |
| `ZOHO_CLIENT_ID` | — | Zoho OAuth client ID |
| `ZOHO_CLIENT_SECRET` | — | Zoho OAuth client secret |
| `ZOHO_REFRESH_TOKEN` | — | Zoho OAuth refresh token |
| `ZOHO_FROM_EMAIL` | — | Zoho sender address |

## Module

```go
import "github.com/thescaffold/gox-apps-notification/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
