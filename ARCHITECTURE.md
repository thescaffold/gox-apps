# Architecture

## Service boundaries

Each service in `libs/<name>` is an independent Go module. Services own their own database tables and never share a DB connection with another service. Cross-service communication happens via:

- **HTTP** — `gox-packages-core/http.Client` with optional HMAC signing for internal calls
- **Event bus** — `gox-packages-core/events.Bus` (in-process, wildcard pub/sub); the host application wires `Subscriptions` maps at startup

## Module system (Goose)

Every service is a [Goose](https://github.com/awesome-goose/goose) module. The key interfaces are:

| Interface | Purpose |
|-----------|---------|
| `types.Module` | Declares imports (sub-modules), exports (shared services), and declarations (local providers) |
| `types.Bootable` | Implemented by background workers (`Boot(kernel)` called after DI is resolved) |
| `types.Middleware` | HTTP middleware attached to individual routes |

## Database

- ORM: GORM via the goose `sql` module
- Each service calls `sql.Child(&sql.Config{Migrations: [...]})` to own its schema
- Migrations are plain Go structs implementing `sql.Migration`
- Default: SQLite for tests, PostgreSQL for production

## Authentication flow

```
Request
  └─ AuthMiddleware (core/auth) — validates Bearer JWT, sets "auth" in context
       └─ identity/pkg.AuthService — Login, BasicAuth, TokenAuth, HMACAuth
            ├─ ResolveRoles      — role names embedded in JWT claims
            └─ ResolvePermissions — permission keys embedded in JWT claims
```

## Event pipeline (polylog)

```
POST /ingest → PipelineService.Ingest
  └─ Insert Event record
       └─ goroutine: PipelineService.Process
            └─ for each Channel matching event type:
                 └─ PipelineService.Discharge → WebhookSink.Sink (HTTP POST)
```

## Background workers

Workers implement `types.Bootable` and start goroutines from `Boot()`:

| Service | Worker | Schedule |
|---------|--------|---------|
| `health` | `SummaryTask` | Every 10 minutes |
| `audit` | `SummaryBatch` | Every hour (+ `apps.cron.heartbeat.hourly`) |

## Publishing workflow

Each service is published independently. Tag format: `<service>/v<semver>`. The CI workflow verifies build + tests; the publish workflow creates the GitHub Release.

```
git tag capital/v1.0.0 && git push origin capital/v1.0.0
```
