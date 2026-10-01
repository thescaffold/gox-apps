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
| `GROUP_NAME` | — | Templates are looked up by (`key`, `group` = this value) |
| `NOTIFICATION_DEFAULT_PROVIDER` | `mailgun` | `mailgun`, `zoho` or `cloudflare` for email; `termii` or `africastalking` for SMS. `NOTIFICATION_EMAIL_DEFAULT_PROVIDER` / `NOTIFICATION_SMS_DEFAULT_PROVIDER` override it per channel |
| `MAILGUN_BASE_URL` | — | e.g. `https://api.mailgun.net` |
| `MAILGUN_DOMAIN` | — | Mailgun sending domain |
| `MAILGUN_TOKEN` | — | Mailgun private API key |
| `MAILGUN_FROM` | — | Sender address |
| `ZOHO_BASE_URL`, `ZOHO_TOKEN`, `ZOHO_FROM_EMAIL`, `ZOHO_FROM_NAME` | — | Zoho Mail |
| `CLOUDFLARE_BASE_URL`, `CLOUDFLARE_ACCOUNT_ID`, `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_FROM_EMAIL`, `CLOUDFLARE_FROM_NAME` | — | Cloudflare Email Service |

## Delivering one notification and knowing the outcome

The `apps.notification.message.new` event is fire-and-forget. A host that needs
to know whether the message went out calls the service directly:

```go
if svc := app.Service(); svc != nil {           // nil until the module is declared
    err := svc.Deliver(map[string]any{
        "reference": "approval-123-user-9",      // makes a retry safe
        "key": "approval-requested", "userId": "user-9", "subject": "…",
        "channels": []string{"web", "email"},
        "data": map[string]any{"owner": "ada@example.com", "name": "Ada"},
    })
}
```

`Deliver` returns an error unless every requested channel delivered. A reference
that was delivered is not sent twice; one that failed is sent again on the same
log row and its in-app message is not duplicated. `NewAppService(db, provider.NewService(db, hmacKey))`
builds the service without the DI container.

## Module

```go
import "github.com/thescaffold/gox-apps/libs/notification/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
