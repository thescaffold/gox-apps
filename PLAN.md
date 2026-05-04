# gox-packages + gox-apps Migration Plan

> Full port of the scaffold TypeScript monorepo (`ntx-packages` + `ntx-apps`) to Go using the **goose** framework.
>
> **Two efforts — execute in order:**
> 1. **gox-packages** (Phase P) — Port `ntx-packages` foundational libs (apps depend on this)
> 2. **gox-apps** (Phase 0–2) — Port `ntx-apps` (depends on gox-packages being complete)

---

## Reference Paths

| Resource              | Path                                                                                  |
|-----------------------|---------------------------------------------------------------------------------------|
| TS packages source    | `/Users/isaiahiroko/Projects/scaffold/ntx-packages/libs/`                             |
| Go packages dest      | `/Users/isaiahiroko/Projects/scaffold/gox-packages/libs/`                             |
| TS apps source        | `/Users/isaiahiroko/Projects/scaffold/ntx-apps/libs/`                                 |
| Go apps dest          | `/Users/isaiahiroko/Projects/scaffold/gox-apps/libs/`                                 |
| Goose framework       | `/Users/isaiahiroko/Projects/awesome-goose/goose/` (`github.com/awesome-goose/goose`) |
| Host mono repo        | `/Users/isaiahiroko/Projects/origine/cloud/server/src/app.module.ts`                  |

---

## How to Resume

Find the first unchecked `[ ]` box and continue from there.
Phase P must be **fully complete** before resuming Phase 0 app work.
Within each package/app, work top-to-bottom in the order listed.

---

## Part 1 — gox-packages

Mirrors `ntx-packages`. Lives in a separate repo at `/Users/isaiahiroko/Projects/scaffold/gox-packages/`.

### Module Naming

```
github.com/thescaffold/gox-packages-core
github.com/thescaffold/gox-packages-polylog
github.com/thescaffold/gox-packages-blobs
github.com/thescaffold/gox-packages-flags
```

### Key Architecture Decisions

**Response Envelope (CRITICAL — must match TS exactly)**
All NTX API responses follow `{"status","title","message","data","meta"}`.
Provided by `gox-packages-core/response`; replaces bare `output.*` calls throughout all apps.

```go
// {"status":"success","title":"Assets","message":"File created","data":{...},"meta":{...}}
type Envelope struct {
    Status  string `json:"status"`
    Title   string `json:"title,omitempty"`
    Message string `json:"message,omitempty"`
    Data    any    `json:"data"`
    Meta    any    `json:"meta,omitempty"`
}
```

**Event System (pub/sub)**
NestJS `EventEmitter2` wildcard topics → Go in-process wildcard event bus.
`TrackerService.Message(type, props)` → `events.Bus.Publish(type, props)`.
Apps export `Subscriptions map[string]events.EventHandler` from their `index.go`.
The host registers subscriptions with the global EventBus at startup.

Wildcard rules: `*` matches exactly one segment, `#` matches zero or more.
Example: `*.*.*.after-insert` matches `apps.identity.user.after-insert`.

**Context Propagation**
x-ntx-* request headers carry user/workspace/scope/preference (base64 JSON).
A goose middleware parses these into an `NTXContext` struct, stored on the goose Context.
Handler DTOs access it via `context:"ntx"` struct tag.

**i18n (Translation)**
YAML files with dotted keys + Mustache templates.
Path pattern: `translations/{locale}/ntx{/group}/{service}.yaml`
Example path: `packages.core.crud.post.create.success`
  → locale=en, service=core, key path=crud.post.create.success
`i18n.Translate()` falls back: requested locale → English → returns key path if missing.
In-memory cache per service name. Non-blocking: missing YAML never crashes.

**CRUD Resource (replaces `api.Resource[T]` for NTX-compatible apps)**
`crud.CrudResource[E, C, U]` is the Go equivalent of `CrudControllerFactory`.
Provides all 10 endpoints with hooks, morphs, pagination meta, i18n messages, unique checks.

**Permission Scopes**
`auth.PermissionScopeType` (Self/Workspace/Global).
Scope rules are declared per-entity and applied automatically via context.

---

### Phase P0 — Repo Setup

- [x] Create `gox-packages/` directory and `go.work` (links goose + all packages)
- [x] Create `gox-packages/Makefile` (test, build, vet, tidy targets mirroring gox-apps)

---

### Phase P1 — gox-core

**Legend:** `[P]` = package file(s), `[T]` = tests

#### P1.1 — response

- [x] Create `gox-packages/libs/core/go.mod` (`github.com/thescaffold/gox-packages-core`)
- [x] `[P]` `response/response.go` — `Envelope` struct; `Success`, `Created`, `Paginated`,
      `Error`, `NotFound`, `BadRequest`, `Conflict`, `Unauthorized`, `Forbidden`, `InternalServerError`;
      custom `ntxOutput` type returns Envelope directly from `Data()` (no double-wrapping)
- [x] `[T]` `tests/response_test.go` — 19 tests, all passing

#### P1.2 — utils ✅

- [x] `[P]` `utils/random.go` — `RandomDigits(n)`, `Random(n)`, `RandomUsername()`, `RandomReferralCode()`
- [x] `[P]` `utils/uuid.go` — `UUID()`, `Reference(service, maxLen)` (host-pid-random-time pattern)
- [x] `[P]` `utils/string.go` — `UCFirst()`, `TitleCase()`, `MaskEmail()`, `TrimString()`, `Prettify()`
- [x] `[P]` `utils/time.go` — `Now()`, `Format(t, layout)`, time constants (FORMAT.FULL_DATE_TIME etc.)
- [x] `[P]` `utils/money.go` — `ToMajor(minor, decimals)`, `ToMinor(major, decimals)`, `ToInt()`
- [x] `[P]` `utils/types.go` — `KeyValue = map[string]any`, `EntityActionType`, `DefaultRoleTypeName`, `PermissionScopeType`
- [x] `[P]` `utils/validate.go` — `IsEmail()`, `IsPhone()`, `IsNumeric()`, `IsNullOrUndefined()`
- [x] `[T]` `tests/utils_test.go` — 43 tests, all passing

#### P1.3 — context ✅

- [x] `[P]` `context/context.go` — `NTXContext` struct (Method, Source, Sink, TracingID,
      UserID, ClientID, WorkspaceID, Config, User, Client, Workspace, Roles, Permissions, Scope, Preference);
      `Parse(headers)`, `Format(ctx)`, `Get(gooseCtx)`, `Set(gooseCtx, ntx)`
- [x] `[P]` `context/middleware.go` — goose `Middleware`: parses x-ntx-* headers on every request,
      stores NTXContext via `ctx.SetValue(contextKey{}, ntx)`
- [x] `[T]` `tests/context_test.go` — plain headers, base64 fields, round-trip, nil fields

#### P1.4 — security ✅

Matches TS `security.util.ts` exactly so identity app User.secret encryption is interoperable.

- [x] `[P]` `security/security.go` — `ToBase64`/`FromBase64`; `Hash`/`Compare` (bcrypt);
      `Encrypt`/`Decrypt` (AES-256-CBC, `iv_hex:ct_hex` format matching TS exactly);
      `GenerateHmac`/`CompareHmac` (HMAC-SHA256, sorted-key JSON payload);
      `CheckSum` (SHA-256 hex); `MD5`
- [x] `[T]` `tests/security_test.go` — base64 round-trip, bcrypt verify, encrypt/decrypt,
      nondeterministic IV, HMAC sorted-key parity, checksum

#### P1.5 — events ✅

In-process wildcard pub/sub. Replaces `@nestjs/event-emitter` + `EventEmitter2`.

- [x] `[P]` `events/bus.go` — `Bus` struct; `Subscribe(pattern, handler)`;
      `Publish(eventType, payload)` (fan-out, non-blocking via goroutines);
      wildcard matching: `*` = one segment, `#` = zero or more (split on `.`)
- [x] `[P]` `events/tracker.go` — `TrackerService`: `Identify`, `Track`, `Page`, `Message`;
      `Message` lowercases eventType — matches TS `tracker.service.ts` exactly
- [x] `[T]` `tests/events_test.go` — 15 tests: exact match, single/multi `*`, `#`, payload
      delivery, multiple subscribers, TrackerService lowercasing

#### P1.6 — filter ✅

Query param → GORM filter. Mirrors TS `makeFilter()` exactly.

```go
type Result struct {
    Conditions []map[string]any  // OR-able WHERE maps; Like("x") values signal LIKE clause
    Page       int
    PerPage    int
    Columns    []string          // SELECT columns; always prepends "id","created_at"
    Relations  []string          // PRELOAD names
    DateFrom   string            // ISO-8601 for createdAt BETWEEN
    DateTo     string
}
func MakeFilter(queries map[string]string, searchable []string) Result
```

Supports: `query=` (LIKE across searchable), `page=`, `perPage=` (default 12),
`columns=a,b`, `relations=x,y`, `from=`/`to=` (createdAt range), dotted keys (`user.name=foo`),
LIKE+field-filter combination, reserved key stripping.

- [x] `[P]` `filter/filter.go` — `MakeFilter()`, `Result`, `Like` type, `Offset()`,
      dotted-notation nested filter builder, `recursiveMerge`
- [x] `[T]` `tests/filter_test.go` — 18 tests: defaults, pagination, columns, relations,
      LIKE search, field filters, dotted keys, combined, date range

#### P1.7 — crud ✅

Go equivalent of `CrudControllerFactory`. Integrates response, i18n, filter, scope.

```go
type Config[E, C, U any] struct {
    Entity     sql.Entity[E]
    Name       string              // used in i18n: "User", "List", etc.
    Searchable []string
    Unique     func(C) []map[string]any  // constraints to check before create/update
    Hooks      map[CrudActionType]HookFn
    Morphs     map[CrudActionType]MorphFn
}
type CrudResource[E, C, U any] struct { ... }
func (r *CrudResource[E,C,U]) Hydrate(cfg Config[E,C,U])
```

Endpoints (all return `types.Output` with Envelope body):
- `Create(dto)` — POST /; unique check; hooks; returns 201 with entity
- `CreateIgnoreDuplicate(dto)` — PUT /; find-or-create
- `List(dto)` — GET /; makeFilter; pagination meta; hooks
- `Get(dto)` — GET /:id; relations; 404 if missing
- `UpdatePatch(dto)` — PATCH /:id; unique check; hooks
- `UpdatePut(dto)` — PUT /:id; full replace
- `Upsert(dto)` — POST /upsert?update=true; updateOrCreate / findOrCreate
- `Delete(dto)` — DELETE /:id; soft delete; 404 if missing
- `Metrics(dto)` — GET /metrics; count with filters
- `FindRelatives(dto)` — GET /:id/:relative; load relation
- `FindByType(dto)` — GET /find/:type; latest/oldest by createdAt
- `FindByIds(dto)` — POST /ids; IN query with pagination

- [x] `[P]` `crud/types.go` — `HookFn`, `MorphFn`, `Config[E,C,U]`, hook name constants
- [x] `[P]` `crud/dto.go` — all DTO structs for each endpoint (param/query/json/context tags)
- [x] `[P]` `crud/middleware.go` — `QueriesMiddleware` stores full query map in context
- [x] `[P]` `crud/resource.go` — `CrudResource[E,C,U]`, `Hydrate()`, all 12 endpoint methods, helpers
- [x] `[T]` `tests/crud_test.go` — full CRUD flow, pagination, unique check, hook execution, 404 paths (26 tests, all pass)

#### P1.8 — auth ✅

JWT + permission scopes. Used by identity app and host AuthMiddleware.

- [x] `[P]` `auth/jwt.go` — `Sign`/`Verify` HS256; expiry=0 omits exp claim
- [x] `[P]` `auth/scope.go` — `ExpandScopeObject`, `GetScopeValue` mirroring TS exactly
- [x] `[P]` `auth/middleware.go` — `AuthMiddleware`; reads Bearer header, 401 + chain stop on failure, stores claims at "auth"
- [x] `[T]` `tests/auth_test.go` — 14 tests, all pass

#### P1.9 — i18n ✅

YAML translation with dotted key resolution and Mustache rendering.

- [x] `[P]` `i18n/loader.go` — `Loader` interface; `FileLoader`; `HTTPLoader` with in-memory cache
- [x] `[P]` `i18n/i18n.go` — `Translate`, `getValueByDottedString`, service + path cache, EN fallback
- [x] `[P]` `i18n/render.go` — `RenderHelpers` map with `cbroglie/mustache` LambdaFunc wrappers for all helpers
- [x] `[T]` `tests/i18n_test.go` — 14 tests (key resolution, locale, fallback, Mustache, lambdas), all pass

#### P1.10 — http ✅

Retry HTTP client for inter-service calls.

- [x] `[P]` `http/client.go` — `Client`; `Internal` (HMAC headers); `External`; shared retry loop; `toStringMap` helper
- [x] `[T]` `tests/http_test.go` — 9 tests (200/404, POST body, query params, HMAC headers, retry success, retry exhaustion, unreachable host), all pass

#### P1.11 — ws (stub) ✅

- [x] `[P]` `ws/ws.go` — `WsService` no-op stub; `Emit` and `Bind` return nil
- [x] `[T]` `tests/ws_test.go` — 2 smoke tests, pass

#### P1.12 — CoreModule ✅

Single importable goose module that wires all core services.

- [x] `[P]` `module/core.module.go` — `CoreModule`; declares EventBus, TrackerService, http.Client, WsService, context.Middleware
- [x] `[T]` `tests/core_module_test.go` — 5 DI wiring smoke tests, all pass

---

### Phase P2 — gox-polylog ✅

Thin analytics wrapper. Delegates to gox-core EventBus by default.

- [x] `events/events.service.go` — `EventsService` wrapping `TrackerService`; `toKV` coercion helper
- [x] `polylog.module.go` — `PolylogModule.Register`; declares Bus + TrackerService + EventsService
- [x] `index.go` — re-exports `EventsService`
- [x] `tests/polylog_test.go` — 10 tests (Identify/Track/Message/lowercase/nil-safe, module wiring), all pass

---

### Phase P3 — gox-blobs ✅

File storage abstraction used by the `blobs` gox-app.

- [x] `files/provider.go` — `StorageProvider` interface
- [x] `files/local.go` — `LocalProvider` (filesystem, URL `/storage/{bucket}/{name}`)
- [x] `files/s3.go` — `S3Provider` stub (returns not-implemented error)
- [x] `files/files.service.go` — `FilesService` delegating to provider
- [x] `blobs.module.go` — `BlobsModule.Register`; selects provider by config
- [x] `index.go` — re-exports `BlobsModule`, `FilesService`, `StorageProvider`
- [x] `tests/blobs_test.go` — 10 tests (local upload/download, S3 stub errors, module wiring), all pass

---

### Phase P4 — gox-flags ✅

Feature flag management. Used by the `flags` gox-app.

- [x] `flag/flag.service.go` — `FlagService`: `Register`, `Log` (with limit), `Status`, `Limit`; env-aware in-memory
- [x] `flags.module.go` — `FlagsModule.Register`
- [x] `index.go` — re-exports `FlagsModule`, `FlagService`, `FlagDef`, opts types
- [x] `tests/flags_test.go` — 13 tests (register, disabled, env filter, log, limit, module wiring), all pass

---

## Part 2 — gox-apps

### Updated Conventions

#### Go Module Naming

```
github.com/thescaffold/gox-apps-<app>
```

#### Directory Layout (per app)

```
libs/<app>/
├── go.mod
├── go.sum
├── app/
│   ├── app.module.go
│   ├── app.controller.go
│   ├── app.service.go
│   ├── app.routes.go
│   ├── app.dto.go
│   ├── index.go            # Name, AppModule, Migrations, Seeders, Subscriptions, DefaultPermissions
│   └── <resource>/
│       ├── <res>.entity.go
│       ├── <res>.service.go
│       ├── <res>.controller.go
│       ├── <res>.routes.go
│       ├── <res>.dto.go
│       └── <res>.module.go
├── migrations/
└── tests/
```

#### Response Format

All controller handlers return the NTX envelope via gox-core:

```go
import "github.com/thescaffold/gox-packages-core/response"

return response.Success(entity, "Assets", "File created", nil)
return response.Error("Assets", "File not found", 404)
return response.Paginated(items, page, perPage, total, "Assets", "Files listed")
```

#### Subscriptions Pattern

```go
// in app/index.go:
import "github.com/thescaffold/gox-packages-core/events"

var Subscriptions = map[string]events.EventHandler{
    "*.*.*.after-insert": func(event string, payload any) { ... },
    "apps.figs.message.new": func(event string, payload any) { ... },
}
```

#### CRUD Controllers

Use `crud.CrudResource[E, C, U]` from gox-core (not goose's `api.Resource[T]`):

```go
import "github.com/thescaffold/gox-packages-core/crud"

type ListController struct {
    crud.CrudResource[List, CreateListDto, UpdateListDto]
    entity *ListEntity `inject:""`
}
func (c *ListController) OnRegister() {
    c.Hydrate(crud.Config[List, CreateListDto, UpdateListDto]{
        Entity:     c.entity,
        Name:       "list",
        Searchable: []string{"code", "value", "key"},
    })
}
```

Routes are unchanged: `router.Resource("list", ListController{}).All()`

#### Context in Handlers

The CoreModule middleware parses x-ntx-* headers automatically.
Access via DTO tag `context:"ntx"` which resolves to `NTXContext`:

```go
type CreateDto struct {
    Name    string      `json:"name"`
    Ctx     NTXContext  `context:"ntx"`
}
```

---

### Phase 0 — Setup (gox-apps)

- [x] Confirm goose framework version (`v0.0.6`)
- [x] Confirm Go version (`1.25.3`)
- [x] Create `gox-apps/libs/` directory structure
- [x] Confirm module path prefix (`github.com/thescaffold/gox-apps-*`)
- [x] Write `Makefile`
- [ ] Add `gox-packages-core` to `gox-apps/go.work` once Phase P1 is complete

---

### Phase 1 — Apps

**Legend:** `[E]` Entity · `[M]` Migration · `[S]` Service · `[C]` Controller · `[R]` Routes · `[Mod]` Module · `[Idx]` index.go · `[T]` Tests · `[U]` Update (post gox-core)

---

### 1. statics ✅ → update required

Initial implementation is complete. Must be updated after gox-core is ready.

- [x] Create `libs/statics/go.mod`
- [x] `[E]` `app/list/list.entity.go`
- [x] `[M]` `migrations/create_statics_lists.go` + `seed_statics_lists_data.go`
- [x] `[S]` `app/list/list.service.go`
- [x] `[C]` `app/list/list.controller.go`
- [x] `[R]` `app/list/list.routes.go`
- [x] `[Mod]` `app/list/list.module.go`
- [x] `[S]` `app/app.service.go`
- [x] `[C]` `app/app.controller.go`
- [x] `[R]` `app/app.routes.go`
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go`
- [x] `[T]` `tests/list_service_test.go` — all passing
- [x] `[U]` Swap `api.Resource[List]` → `crud.CrudResource[List, CreateListDto, UpdateListDto]`
- [x] `[U]` Swap `output.*` → `response.*` envelope in AppController
- [x] `[U]` Add `CoreModule` to `app.module.go` imports
- [x] `[U]` Update `index.go` Subscriptions with typed `events.EventHandler` (statics has none, add empty map)
- [x] `[U]` Re-run tests — all passing

---

### 2. audit ✅ → update required

- [x] Create `libs/audit/go.mod`
- [x] `[E]` `app/log/log.entity.go`
- [x] `[M]` `migrations/create_audit_logs.go`
- [x] `[S]` `app/log/log.service.go`
- [x] `[C]` `app/log/log.controller.go`
- [x] `[R]` `app/log/log.routes.go`
- [x] `[Mod]` `app/log/log.module.go`
- [x] `[S]` `app/app.service.go` — `HandleEvent(payload)`
- [x] `[C]` `app/app.controller.go`
- [x] `[R]` `app/app.routes.go`
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go`
- [x] `[T]` `tests/log_service_test.go` — all passing
- [x] `[U]` Swap `api.Resource[Log]` → `crud.CrudResource[Log, CreateLogDto, UpdateLogDto]`
- [x] `[U]` Swap `output.*` → `response.*` envelope
- [x] `[U]` Wire `Subscriptions`: replace comment-style stubs with typed `events.EventHandler` map;
      `HandleEvent` becomes the EventBus handler body for `*.*.*.after-insert/update/delete`
- [x] `[U]` Add `CoreModule` to imports
- [x] `[U]` Re-run tests — all passing

---

### 3. assets ✅ → update required

- [x] Create `libs/assets/go.mod`
- [x] `[E]` `app/file/file.entity.go`
- [x] `[M]` `migrations/create_assets_files.go`
- [x] `[S]` `app/file/file.service.go`
- [x] `[C]` `app/file/file.controller.go`
- [x] `[R]` `app/file/file.routes.go`
- [x] `[Mod]` `app/file/file.module.go`
- [x] `[S]` `app/app.service.go`
- [x] `[C]` `app/app.controller.go`
- [x] `[R]` `app/app.routes.go`
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go`
- [x] `[T]` `tests/file_service_test.go` — all passing
- [x] `[U]` Swap `output.*` → `response.*` envelope in AppController
- [x] `[U]` Add `CoreModule` to imports
- [x] `[U]` Update `index.go` Subscriptions with empty typed `events.EventHandler` map
- [x] `[U]` Re-run tests — all passing

---

### 4. figs (in progress)

Full file-transformation pipeline (converter/mapper/store/validator provider pattern).
Triggered via `apps.figs.message.new` subscription.

- [x] Create `libs/figs/go.mod`
- [x] `[E]` `app/file/file.entity.go` — `FigsFiles` (name, input/output/meta jsonb, url, raw, tags, status)
- [x] `[S]` `app/file/file.service.go`
- [x] `[P]` `pkg/converter/` — `Service`, `CSVProvider`, `ExcelProvider` (stub), `PDFProvider` (stub)
- [x] `[P]` `pkg/mapper/` — `Service`, `DataProvider`, `HTMLProvider`
- [x] `[P]` `pkg/store/` — `Service`, `LocalProvider`, `ObjectProvider` (stub)
- [x] `[P]` `pkg/validator/` — `Service`, `JSONProvider` (stub), `YAMLProvider` (stub)
- [x] `[S]` `app/app.service.go` — `Init()`, `Process()`, `SaveFile()`, `UpdateFile()`
- [x] `[C]` `app/app.controller.go` — `GET /` health, `POST /:name` pipeline
- [x] `[R]` `app/app.routes.go`
- [x] `[Mod]` `app/file/file.module.go`
- [x] `[M]` `migrations/create_figs_files.go`
- [x] `[Mod]` `app/app.module.go` — imports sql.Child + file module + pkg services + ROUTES
- [x] `[Idx]` `app/index.go` — `Name="figs"`, `Migrations`, `Subscriptions` (apps.figs.message.new → Init+Process+UpdateFile)
- [x] `[U]` Swap `output.*` → `response.*` envelope (after gox-core)
- [x] `[T]` `tests/file_service_test.go` — all passing

---

### 5. cache

**TS source:** `libs/cache/src/` · **Resources:** `list`
**Note:** `CacheList` stores serialized key-value cache entries in DB. Distinct from goose `kv` module
(which is for runtime caching); this is for persistent user-facing cached data.

- [x] Create `libs/cache/go.mod`
- [x] `[E]` `app/list/list.entity.go` — `CacheLists` table (key, value jsonb, group, ttl, status)
- [x] `[M]` `migrations/create_cache_lists.go`
- [x] `[S]` `app/list/list.service.go` — `Get(key, group)`, `Set(key, value, group, ttl)`, `Del(key, group)`, `Flush(group)`
- [x] `[C]` `app/list/list.controller.go` — `crud.CrudResource[List, ...]`
- [x] `[R]` `app/list/list.routes.go`
- [x] `[Mod]` `app/list/list.module.go`
- [x] `[C]` `app/app.controller.go` — `GET /` health; `GET /get` get; `POST /set` set; `DELETE /del` del; `DELETE /flush` flush
- [x] `[R]` `app/app.routes.go`
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go`
- [x] `[T]` `tests/list_service_test.go` — all passing

---

### 6. blobs

**TS source:** `libs/blobs/src/` · **Resources:** `file`, `page`
**Note:** Uses `gox-packages-blobs` (`FilesService`) for actual storage; DB tracks metadata.

- [x] Create `libs/blobs/go.mod`
- [x] `[E]` `app/file/file.entity.go` — `BlobsFiles` (name, type, size, bucket, url, tags, status)
- [x] `[M]` `migrations/create_blobs_files.go`
- [x] `[S]` `app/file/file.service.go` — wraps `gox-packages-blobs` FilesService for upload/download
- [x] `[C]` `app/file/file.controller.go` — `crud.CrudResource[File, ...]`
- [x] `[R]` `app/file/file.routes.go`
- [x] `[Mod]` `app/file/file.module.go`
- [x] `[E]` `app/page/page.entity.go` — `BlobsPages` (fileId, index, raw, status, meta)
- [x] `[M]` `migrations/create_blobs_pages.go`
- [x] `[S]` `app/page/page.service.go`
- [x] `[C]` `app/page/page.controller.go` — `crud.CrudResource[Page, ...]`
- [x] `[R]` `app/page/page.routes.go`
- [x] `[Mod]` `app/page/page.module.go`
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go`
- [x] `[T]` `tests/file_service_test.go` — all passing

---

### 7. health

**TS source:** `libs/health/src/` · **Resources:** `log`, `service`, `summary`

- [x] Create `libs/health/go.mod`
- [x] `[E]` `app/log/log.entity.go` — `HealthLogs` (serviceId, state, meta jsonb)
- [x] `[M]` `migrations/create_health_logs.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `log`
- [x] `[E]` `app/service/service.entity.go` — `HealthServices` (name, desc, type, state, status)
- [x] `[M]` `migrations/create_health_services.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `service`
- [x] `[E]` `app/summary/summary.entity.go` — `HealthSummaries` (serviceId, type, received, measure, note)
- [x] `[M]` `migrations/create_health_summaries.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `summary`
- [x] `[S]` `app/app.service.go` — `Check(serviceId)` → logs result
- [x] `[C]` `app/app.controller.go` — health endpoint + `POST /check/:id`
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go` — export `Subscriptions` (3 events)
- [x] `[T]` tests — all passing

---

### 8. cron

**TS source:** `libs/cron/src/` · **Resources:** `job`, `log`
**Note:** Uses goose `cron` module for scheduling. Jobs emit events consumed by other apps.

- [x] Create `libs/cron/go.mod`
- [x] `[E]` `app/job/job.entity.go` — wraps `gocron.CronJob` (goose manages cron_jobs table)
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `job`
- [x] `[E]` `app/log/log.entity.go` — wraps `gocron.CronLog` (goose manages cron_logs table)
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `log`
- [x] `[S]` `app/app.service.go` — `Register/Select/Log` via goose `Cron` service; `PublishHeartbeat` via EventBus
- [x] `[C]` `app/app.controller.go`
- [x] `[Mod]` `app/app.module.go` — imports `gocron.NewModule(DefaultConfig(), Jobs, true)`
- [x] `[Idx]` `app/index.go` — `Jobs []*gocron.CronHandler` (6 heartbeats); `Subscriptions` (hourly)
- [x] `[T]` tests — all passing
- **Note:** No separate migrations — goose `cron.NewModule` creates `cron_jobs`+`cron_logs` tables

---

### 9. queue

**TS source:** `libs/queue/src/` · **Resources:** `queue`, `job`, `log`
**Note:** Uses goose `queues` module for job processing.

- [x] Create `libs/queue/go.mod`
- [x] `[E]` `app/queue/queue.entity.go` — wraps `goqueues.QueueQueue`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `queue`
- [x] `[E]` `app/job/job.entity.go` — wraps `goqueues.QueueJob`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `job`
- [x] `[E]` `app/log/log.entity.go` — wraps `goqueues.QueueLog`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `log`
- [x] `[S]` `app/app.service.go` — `Push/Pop/Log` via goose `Queue` service
- [x] `[Mod]` `app/app.module.go` — imports `goqueues.NewModule(nil, nil, true)`
- [x] `[Idx]` `app/index.go` — `Jobs []*goqueues.JobHandler` (empty); `Subscriptions` (hourly)
- [x] `[T]` tests — all passing
- **Note:** No separate migrations — goose `queues.NewModule` creates tables

---

### 10. fuss

**TS source:** `libs/fuss/src/` · **Resources:** `token`, `token-log`, `history`
**Note:** Fuss = "fuzzy session service" — manages one-time and session tokens.

- [x] Create `libs/fuss/go.mod`
- [x] `[E]` `app/token/token.entity.go` — `FussTokens` (userId, clientId, workspaceId, group, service, entityId, entityName, meta jsonb, status)
- [x] `[M]` `migrations/create_fuss_tokens.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `token`
- [x] `[E]` `app/tokenlog/tokenlog.entity.go` — `FussTokenLogs` (tokenId, value, status)
- [x] `[M]` `migrations/create_fuss_token_logs.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `tokenlog`
- [x] `[E]` `app/history/history.entity.go` — `FussHistories` (userId, clientId, workspaceId, type, query, status)
- [x] `[M]` `migrations/create_fuss_histories.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `history`
- [x] `[S]` `app/app.service.go` — `HandleCreate/Update/Delete`, `Recent`, `Search`, `Lookup`
- [x] `[C]` `app/app.controller.go` — Health, Recent, Search, Lookup
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go` — Subscriptions: after-insert/update/delete → HandleCreate/Update/Delete
- [x] `[T]` tests — all passing

---

### 11. controller

**TS source:** `libs/controller/src/` · **Resources:** `route`, `request`

- [x] Create `libs/controller/go.mod`
- [x] `[E]` `app/route/route.entity.go` — `ControllerRoutes` (group, service, type, name, desc, upstream, status)
- [x] `[M]` `migrations/create_controller_routes.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `route`
- [x] `[E]` `app/request/request.entity.go` — `ControllerRequests` (group, service, type, method, url, queries/body/response jsonb, status)
- [x] `[M]` `migrations/create_controller_requests.go`
- [x] `[S]` + `[C]` + `[R]` + `[Mod]` for `request`
- [x] `[S]` `app/app.service.go` — `GetRoutes(limit)`, `GetRoute(path)` → upstream URL
- [x] `[C]` `app/app.controller.go` — Health, GetRoutes, GetRoute
- [x] `[Mod]` `app/app.module.go`
- [x] `[Idx]` `app/index.go`
- [x] `[T]` tests — all passing

---

### 12. common ✅

**TS source:** `libs/common/src/` · **Resources:** `ip`, `project`, `project-type`, `rate`, `rate-log`, `tag`, `tag-type`

- [x] Create `libs/common/go.mod`
- [x] `[E]` `app/ip/ip.entity.go` — `CommonIps` (address, country, region, city, isp, meta jsonb)
- [x] `[M]` `migrations/create_common_ips.go`
- [x] Full stack (S+C+R+Mod) for `ip`
- [x] `[E]` `app/project/project.entity.go` — `CommonProjects` (userId, clientId, workspaceId, typeId, name, desc, meta jsonb, status)
- [x] `[M]` `migrations/create_common_projects.go`
- [x] Full stack for `project` and `project_type`
- [x] `[E]` `app/rate/rate.entity.go` — `CommonRates` (name, value, currency, type, status)
- [x] Full stack for `rate` and `rate_log`
- [x] `[E]` `app/tag/tag.entity.go` — `CommonTags` (groupName, serviceName, entityName, entityId, typeId, status)
- [x] Full stack for `tag` and `tag_type`
- [x] `[S]` `app/app.service.go` — `GetCurrency(currency)`, `GetLocation(ipValue)`
- [x] `[Mod]` `app/app.module.go` — 7 migrations, all 7 resource modules
- [x] `[Idx]` `app/index.go`
- [x] `[T]` tests — all passing

---

### 13. flags ✅

**TS source:** `libs/flags/src/` · **Resources:** `flag`, `flag-log`, `environment`, `environment-type`

- [x] Create `libs/flags/go.mod`
- [x] `[E]` `app/environmenttype/environmenttype.entity.go` — `FlagEnvironmentTypes` (category, name, desc, detail, thumbnailUrl, bannerUrl, tags jsonb, meta jsonb)
- [x] `[M]` `migrations/create_flag_environment_types.go`
- [x] Full stack for `environmenttype`
- [x] `[E]` `app/environment/environment.entity.go` — `FlagEnvironments` (userId, clientId, workspaceId, typeId, name, desc, meta jsonb, status)
- [x] `[M]` `migrations/create_flag_environments.go`
- [x] Full stack for `environment`
- [x] `[E]` `app/flag/flag.entity.go` — `FlagFlags` (userId, clientId, workspaceId, environmentId, name, limit, priority, level, meta jsonb, status)
- [x] `[M]` `migrations/create_flag_flags.go`
- [x] Full stack for `flag`
- [x] `[E]` `app/flaglog/flaglog.entity.go` — `FlagFlagLogs` (flagId, limit, meta jsonb, status)
- [x] `[M]` `migrations/create_flag_flag_logs.go`
- [x] Full stack for `flaglog`
- [x] `[Mod]` `app/app.module.go` — 4 migrations
- [x] `[Idx]` `app/index.go`
- [x] `[T]` tests

---

### 14. forms ✅

**TS source:** `libs/forms/src/` · **Resources:** `form`, `form-field`, `form-type`, `form-log`

- [x] Create `libs/forms/go.mod`
- [x] `[E]` `app/formtype/formtype.entity.go` — `FormFormTypes` (name, desc, detail, thumbnailUrl, bannerUrl, tags jsonb, meta jsonb)
- [x] `[M]` `migrations/create_form_form_types.go`
- [x] Full stack for `formtype`
- [x] `[E]` `app/form/form.entity.go` — `FormForms` (userId, clientId, workspaceId, typeId, key, name, desc, meta jsonb, status)
- [x] `[M]` `migrations/create_form_forms.go`
- [x] Full stack for `form`
- [x] `[E]` `app/formfield/formfield.entity.go` — `FormFormFields` (formId, key, name, type, desc, placeholder, disabled bool, required bool, defaultValue, meta jsonb, status)
- [x] `[M]` `migrations/create_form_form_fields.go`
- [x] Full stack for `formfield`
- [x] `[E]` `app/formlog/formlog.entity.go` — `FormFormLogs` (formId, formFieldId, key, value, meta jsonb, status)
- [x] `[M]` `migrations/create_form_form_logs.go`
- [x] Full stack for `formlog`
- [x] `[Mod]` `app/app.module.go` — 4 migrations
- [x] `[Idx]` `app/index.go`
- [x] `[T]` tests

---

### 15. notification ✅

**TS source:** `libs/notification/src/` · **Resources:** `template`, `rule`, `message`, `log`

- [x] Create `libs/notification/go.mod`
- [x] `[E]` `app/template/template.entity.go` — `NotificationTemplates` (group, key, email text, sms text, web text, mobile text, status)
- [x] `[M]` `migrations/create_notification_templates.go`
- [x] Full stack for `template`
- [x] `[E]` `app/rule/rule.entity.go` — `NotificationRules` (group, key, rules jsonb, status)
- [x] Full stack for `rule`
- [x] `[E]` `app/message/message.entity.go` — `NotificationMessages` (reference, userId, clientId, workspaceId, key, subject, channel, message jsonb, timestamps, type, priority, theme, scope, position, status)
- [x] Full stack for `message`
- [x] `[E]` `app/log/log.entity.go` — `NotificationLogs` (reference, userId, clientId, workspaceId, key, subject, channels jsonb, data jsonb, message jsonb, meta jsonb, timestamps, type, priority, theme, scope, position, status, templateId, ruleId)
- [x] Full stack for `log`
- [x] `[Mod]` `app/app.module.go` — 4 migrations
- [x] `[Idx]` `app/index.go`
- [x] `[T]` tests

---

### 16. polylog ✅

**TS source:** `libs/polylog/src/` · **Resources:** `source-type`, `source`, `sink-type`, `sink`, `channel`, `event`, `event-log`

- [x] Create `libs/polylog/go.mod`
- [x] `[E]` + `[M]` + full stack for `sourcetype`, `source`, `sinktype`, `sink`, `channel`
- [x] `[E]` `app/event/event.entity.go` — `PolylogEvents` (userId, clientId, workspaceId, entityId, entityName, category, type, reference, version, payload jsonb, status)
- [x] `[M]` `migrations/create_polylog_events.go`
- [x] Full stack for `event` and `eventlog`
- [x] `[Mod]` `app/app.module.go` — 7 migrations
- [x] `[Idx]` `app/index.go`
- [x] `[T]` tests

---

### 17. bridge ✅

**TS source:** `libs/bridge/src/` · **Resources:** `license`, `license-type`, `plan-type`, `preference`, `webhook`, `webhook-log`
**Note:** Job handlers subscribe to capital payment events to manage licenses.

- [x] Create `libs/bridge/go.mod`
- [x] `[E]` `app/license/license.entity.go` — `BridgeLicenses`
- [x] `[M]` `migrations/create_bridge_licenses.go`
- [x] Full stack for `licensetype`, `license`, `plantype`, `preference`, `webhook`, `webhooklog`
- [x] `[Mod]` `app/app.module.go` — 6 migrations
- [x] `[Idx]` `app/index.go` — `Subscriptions`: `apps.cron.heartbeat.daily`
- [x] `[T]` tests — `go test ./libs/bridge/...` passes

---

### 18. capital ✅

**TS source:** `libs/capital/src/` · **Resources:** `account`, `payment`, `payment-log`, `plan`, `plan-type`, `provider`, `rate`, `transaction`, `usage`, `voucher`, `voucher-type`
**Note:** 11 resources; monetary values use `int` (minor units); `go build` and `go test` pass.

- [x] Create `libs/capital/go.mod`
- [x] Full stack for `account`, `provider`, `rate`, `plantype`, `plan`, `payment`, `paymentlog`, `transaction`, `usage`, `vouchertype`, `voucher`
- [x] `[Mod]` `app/app.module.go` — 11 migrations
- [x] `[Idx]` `app/index.go` — `Subscriptions`: `apps.cron.heartbeat.daily`
- [x] `[T]` tests — `go test ./libs/capital/...` passes

---

### 19. identity ✅

**TS source:** `libs/identity/src/`
**Resources:** `attribute`, `client`, `client-log`, `device`, `device-log`, `device-session`, `invite`, `permission`, `permission-type`, `provider`, `provider-log`, `role`, `role-type`, `token`, `user`, `user-client-workspace`, `workspace`
**Note:** 17 resources, 17 migrations. Auth service / OAuth stubs / JWT wiring deferred to Phase 2.

- [x] Create `libs/identity/go.mod`
- [x] Full stack for `user`, `client`, `clientlog`, `workspace`, `userclientworkspace`
- [x] Full stack for `roletype`, `role`, `permissiontype`, `permission`
- [x] Full stack for `token`, `device`, `devicelog`, `devicesession`
- [x] Full stack for `attribute`, `invite`, `provider`, `providerlog`
- [x] `[Mod]` `app/app.module.go` — 17 migrations
- [x] `[Idx]` `app/index.go` — `Subscriptions`: `apps.cron.heartbeat.hourly`, `*.*.*.after-insert`; re-exports `AuthModule`, `AuthMiddleware`
- [x] `[S]` `pkg/auth.service.go` — `Login`, `RefreshToken`, `Logout`, `ValidateToken`, `RegisterUser`
- [x] `[C]` `pkg/auth.controller.go` — `POST /auth/login`, `POST /auth/logout`, `POST /auth/refresh`, `GET /auth/me`
- [x] `[C]` `pkg/oauth.controller.go` — `GET /oauth/:provider`, `GET /oauth/:provider/callback` (stubs)
- [x] `[Mod]` `pkg/auth.module.go` — `AuthModule` wiring auth routes + `AuthMiddleware` (Bearer JWT via `gox-core/auth`)
- [x] `[T]` tests — `go test ./libs/identity/...` passes; includes `ValidateToken` round-trip + invalid token tests

---

## Phase 2 — Integration

- [x] `gox-packages/go.work` workspace linking all 4 packages + goose
- [x] `gox-apps/go.work` updated to reference gox-packages modules
- [x] All `go build` pass in both repos (all 19 gox-apps libs + 4 gox-packages libs)
- [x] All `go test` pass in both repos (all 19 gox-apps libs + 4 gox-packages libs)
- [x] `INTEGRATION.md` — module imports, migration ordering, subscriptions, cron jobs, auth middleware, RBAC seed
- [x] Migration ordering guide (dependency order for combined migration slice in host)

---

## TypeScript → Go Translation Reference (Updated)

| TypeScript (`ntx-packages` / `ntx-apps`)          | Go (`gox-packages` / `gox-apps`)                                  |
|----------------------------------------------------|-------------------------------------------------------------------|
| `success(data, title, message)`                   | `response.Success(data, title, message, nil)`                     |
| `error(title, message, 404)`                      | `response.Error(title, message, 404)`                             |
| `CrudControllerFactory(createDto, updateDto, E)`  | `crud.CrudResource[E, C, U]` from gox-core                        |
| `BaseRepository<T>.withContext(ctx).find()`       | `sql.Entity[T]` + gox-core scope filters in service layer         |
| `TrackerService.message(type, props)`             | `events.Bus.Publish(type, props)`                                 |
| `@GetContext() context`                           | `context:"ntx"` DTO tag → `NTXContext`                            |
| `makeFilter(queries, searchable)`                 | `filter.MakeFilter(queries, searchable)`                          |
| `LanguageService.translate(path, media, pref, d)` | `i18n.Translate(path, loader, pref, data)`                        |
| `encrypt(text, key)` / `decrypt(cipher, key)`     | `security.Encrypt(text, key)` / `security.Decrypt(cipher, key)`   |
| `subscriptions = { pattern: handler }`            | `var Subscriptions = map[string]events.EventHandler{...}`         |
| `@nestjs/event-emitter` wildcard `*`              | `events.EventBus` with `*`/`#` wildcard matching                  |
| `x-ntx-*` request headers                        | parsed by `context.Middleware`, injected via `context:"ntx"` tag  |
| `formatHeaders(ctx, source, sink)`                | `context.FormatHeaders(ctx, source, sink)`                        |
| `paths: [...]` in `index.ts`                      | YAML loader path config in `i18n.FileLoader`                      |
| `defaultRolesNPermissions` in `index.ts`          | `DefaultPermissions` map exported from `app/index.go`             |
| `jobs: [...]` in `index.ts`                       | `Jobs []JobDef` exported from `app/index.go` for host to register |

---

## Notes

### Encryption Interoperability
TS `encrypt`/`decrypt` uses AES-256-GCM. Go `security.Encrypt`/`Decrypt` must use the same cipher,
key size, nonce strategy, and output encoding to ensure existing DB ciphertext is readable from Go.
Read and test the exact TS implementation before coding the Go version.

### Wildcard Event Matching
TS `EventEmitter2` uses `.` as delimiter. The Go EventBus must match:
- `apps.cron.heartbeat.weekly` matches subscription `apps.cron.heartbeat.*`
- `apps.identity.user.after-insert` matches `*.*.*.after-insert`
- `#` (hash) matches zero or more segments at the end (as in AMQP topics)

### RBAC DefaultPermissions
Each app's `index.go` exports `DefaultPermissions map[string][]string` (role → permission list).
The host merges these maps when setting up its RBAC configuration.

### i18n YAML Files
YAML files live at `translations/{locale}/ntx{/group}/{service}.yaml`.
Group is `nil` for core (`packages.core.*` → `translations/en/ntx/core.yaml`).
Group is the first segment for apps (`apps.identity.*` → `translations/en/ntx/apps/identity.yaml`).
The loader can read from filesystem (local dev) or HTTP CDN (production) — same as TS behavior.
Missing keys fall back to key path string (never crash), so i18n is always non-blocking.

### gox-apps go.work After Phase P
Once gox-packages-core is published (or available locally via go.work), add to `gox-apps/go.work`:
```
use ../../gox-packages/libs/core
use ../../gox-packages/libs/polylog
use ../../gox-packages/libs/blobs
use ../../gox-packages/libs/flags
```
