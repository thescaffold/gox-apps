# Integration Guide

How a host application imports and wires up `gox-apps` + `gox-packages`.

---

## 1. Workspace setup

Both repos expose their modules via `go.work`. A host that imports from either repo should
add the relevant `use` directives to its own `go.work`, or use the published module versions.

```
// host/go.work
go 1.25.3

use .
use ../../scaffold/gox-packages/libs/core
use ../../scaffold/gox-apps/libs/identity
use ../../scaffold/gox-apps/libs/capital
// … add other app libs as needed
```

For local development with goose:
```
use ../../awesome-goose/goose
```

---

## 2. Registering an app module

Each gox-app exposes an `AppModule` in its `app` package. Import it into your host's root module:

```go
import (
    identityapp "github.com/thescaffold/gox-apps-identity/app"
    capitalapp  "github.com/thescaffold/gox-apps-capital/app"
)

type HostModule struct{}

func (m *HostModule) Imports() []types.Module {
    return []types.Module{
        &identityapp.AppModule{},
        &capitalapp.AppModule{},
        // …
    }
}
```

---

## 3. Migration ordering

Each app's `AllMigrations` slice is in dependency order within that app. When combining
migrations from multiple apps in the host, observe this ordering:

| Order | App       | Reason |
|-------|-----------|--------|
| 1     | identity  | Users, clients, workspaces — everything else FK's to these |
| 2     | capital   | Plans reference nothing from other apps |
| 3     | bridge    | Licenses reference capital plans (conceptually) |
| 4     | flags     | Environments/flags reference nothing |
| 5     | forms     | Forms reference nothing |
| 6     | notification | Messages/logs reference nothing |
| 7     | polylog   | Events reference nothing |
| 8     | statics   | Independent |
| 9     | audit     | Independent |
| 10    | assets    | Independent |
| 11    | figs      | Independent |
| 12    | cache     | No migrations |
| 13    | health    | No migrations |
| 14    | cron      | No migrations |
| 15    | queue     | No migrations |
| 16    | fuss      | No migrations |
| 17    | controller | No migrations |
| 18    | common    | No migrations |

Combined migration slice example:

```go
import (
    identityapp      "github.com/thescaffold/gox-apps-identity/app"
    capitalapp       "github.com/thescaffold/gox-apps-capital/app"
    bridgeapp        "github.com/thescaffold/gox-apps-bridge/app"
    flagsapp         "github.com/thescaffold/gox-apps-flags/app"
    formsapp         "github.com/thescaffold/gox-apps-forms/app"
    notificationapp  "github.com/thescaffold/gox-apps-notification/app"
    polylogapp       "github.com/thescaffold/gox-apps-polylog/app"
    staticsapp       "github.com/thescaffold/gox-apps-statics/app"
    auditapp         "github.com/thescaffold/gox-apps-audit/app"
    "github.com/awesome-goose/goose/modules/sql"
)

var AllMigrations []sql.Migration

func init() {
    AllMigrations = append(AllMigrations, identityapp.AllMigrations...)
    AllMigrations = append(AllMigrations, capitalapp.AllMigrations...)
    AllMigrations = append(AllMigrations, bridgeapp.AllMigrations...)
    AllMigrations = append(AllMigrations, flagsapp.AllMigrations...)
    AllMigrations = append(AllMigrations, formsapp.AllMigrations...)
    AllMigrations = append(AllMigrations, notificationapp.AllMigrations...)
    AllMigrations = append(AllMigrations, polylogapp.AllMigrations...)
    AllMigrations = append(AllMigrations, staticsapp.AllMigrations...)
    AllMigrations = append(AllMigrations, auditapp.AllMigrations...)
}
```

---

## 4. Event subscriptions

Each app's `Subscriptions` map wires event patterns to handlers. The host registers all of
them with the event bus after bootstrapping:

```go
import (
    "github.com/thescaffold/gox-packages-core/events"
    identityapp "github.com/thescaffold/gox-apps-identity/app"
    capitalapp  "github.com/thescaffold/gox-apps-capital/app"
    cronapp     "github.com/thescaffold/gox-apps-cron/app"
)

func RegisterSubscriptions(bus *events.Bus) {
    for pattern, handler := range identityapp.Subscriptions {
        bus.Subscribe(pattern, handler)
    }
    for pattern, handler := range capitalapp.Subscriptions {
        bus.Subscribe(pattern, handler)
    }
    // … repeat for every app
}
```

Subscription patterns used across all apps:

| Pattern                      | Subscribers |
|------------------------------|-------------|
| `apps.cron.heartbeat.hourly` | flags, forms, notification, polylog, identity |
| `apps.cron.heartbeat.daily`  | bridge, capital |
| `*.*.*.after-insert`         | polylog, identity |

---

## 5. Cron jobs

The `cron` app exports `Jobs []*gocron.CronHandler`. Register them with the scheduler:

```go
import cronapp "github.com/thescaffold/gox-apps-cron/app"

scheduler.RegisterAll(cronapp.Jobs)
```

---

## 6. Auth middleware (identity)

The identity app's auth module (deferred — see PLAN.md §19) will export `AuthMiddleware`
for the host to inject as a goose middleware:

```go
// When implemented:
import identitypkg "github.com/thescaffold/gox-apps-identity/pkg"

gooseApp.Use(identitypkg.AuthMiddleware)
```

Until then, JWT validation must be handled in the host directly using `gox-packages-core/auth`.

---

## 7. Default permissions (RBAC)

Each app exports `DefaultPermissions map[string][]string`. Merge them in the host to build
the full RBAC seed:

```go
func MergePermissions(apps ...map[string][]string) map[string][]string {
    merged := map[string][]string{}
    for _, perms := range apps {
        for role, keys := range perms {
            merged[role] = append(merged[role], keys...)
        }
    }
    return merged
}

seed := MergePermissions(
    identityapp.DefaultPermissions,
    capitalapp.DefaultPermissions,
    bridgeapp.DefaultPermissions,
    // …
)
```
