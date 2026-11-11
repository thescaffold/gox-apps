# Changelog

All notable changes to gox-apps are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## identity v0.0.26 — other people's accounts are a platform admin's

Security fix (PLAN U26, the same hole Yan's fork closed). **Breaking** for any console that let workspace owners use the generic resources.

- The generic resources over many people's rows — `user`, `token`, `device`, `device-log`, `device-session`, `attribute`, `client`, `client-log`, `invite`, `provider` (its list too), `provider-log`, `user-client-workspace`, `workspace` — and `POST /token/user` need a platform admin (`ROOT_USER_ID`, or a `root` / `global-admin` role). Before, any signed-in person could read, edit or delete another account by id (`PATCH /user/:id` changed its `ref`, `status` or stored a raw `secret`), make themselves owner of anyone's workspace (`POST /user-client-workspace`) and mint tokens for anyone; anyone could list provider rows. A person's own records stay on the account routes.
- Tests: Postgres pools are capped and idle connections expire, so the suite no longer runs a shared Postgres out of clients.

**Check your data:** `IdentityUserClientWorkspaces` rows that make someone an owner of a workspace they did not create, and users whose `ref` or `status` changed unexpectedly.

## Scaffold apps mounted with no middleware decide who may call them (PLAN U26)

Security fix. **Breaking** for anyone who relied on the open routes. Needs core with `coreauth.Caller`, `Admin`, `Internal`, `SecureCaller`, `SecureAdmin` and `AdminWrites`. Every product mounts these apps with `router.Mount("apps/<name>", ...)` and no middleware, so each route now carries its own guard: **public**, **Caller** (a session, acting only on the caller's own data), **Admin** (`ROOT_USER_ID` or a `root` / `global-admin` role; a workspace owner is not one), or **Internal** (a call signed by `core/http.Client.Internal` with `HMAC_KEY`, or an admin's session). Services that call the Internal hooks over HTTP must sign with `Client.Internal`.

- **polylog v0.0.10:** every resource, `config`, `root-source`, `root-sink` and both dashboards (platform-wide counts, open to anyone) need an admin. The webhook routes (`ingest/:id`, `ingest/source/:id`, `ingest/channel/:id`) stay public; `POST /ingest` and `/ingest/batch` stay open for browser SDKs but refuse any item for a source with a signature rule. Dashboards and ingest are documented with typed responses. Hosts whose web client registers `root-source` on every signed-in load (scaffold ngx-common before 2673ed2) now get 403 there.
- **cron v0.0.11, queue v0.0.10, controller v0.0.10, health v0.0.10:** register, pop, push, log, ping and the routes other services call are Internal (anyone could register cron jobs, read other apps' queue payloads, or make the health app poll a URL of their choosing); their resources are admin only. `GET /` and `controller/now` stay public.
- **cache v0.0.9:** every route is admin only (it holds codes and session data).
- **flags v0.0.10:** `register`, `log`, `status`, `limit` are Internal; `environment`, `environment-type`, `flag`, `flag-log` admin only.
- **figs v0.0.11, statics v0.0.10, assets v0.0.9:** writes need an admin; statics lists stay readable by anyone, assets files by a signed-in caller.
- **audit v0.0.14:** `log` is readable within the caller's own workspace and written by an admin; `activities` takes the workspace from the session (mounted bare it always saw none).
- **blobs v0.0.9:** uploads and downloads need a session and stay in the caller's workspace; a missing file is a 404, never the storage error. `file` is workspace-scoped; `page` readable through the caller's files only.
- **fuss v0.0.10:** search, recent and lookup need a session and see only the caller's workspace (search matched every tenant's index); `history` is the caller's own, `token` the workspace's, `token-log` admin only.
- **forms v0.0.9:** forms, fields and submissions belong to their workspace; `one/:id` stays public on purpose (render a shared form, store a submission).
- **notification v0.0.14:** `scope` and `priority` need a session; a subscription is readable and changeable only by the address's owner; `rule` and `template` are admin only; `log` the recipient's own; `message` takes the caller's context so `ScopeToOwner` works, and is an admin's only while a host has not called it. A caller with no workspace sees only the inbox messages that belong to none (it saw every workspace's). A read publishes `apps.notification.inbox.read`. Inbox responses are typed (`Page[InboxItem]`).

## capital v0.0.10 — every route says who may call it

Security fix (PLAN U26). **Breaking** for anyone who relied on the open routes; needs core with `coreauth.Caller/Admin/Internal/AdminWrites`.

- The app routes that act on a plan, payment or card (`debt`, `accrual`, `status`, `user-plan`, `init`, `default-provider`, `pay`, `verify/:type`, `:type/reference/:reference`, `:type/charge/:providerId`, `user-plan/compute-upgrade`), the wallet (`wallet/*`) and `voucher/list` / `voucher/verify/:token` need a session and act on the verified caller only. Before, all answered anonymous callers. `GET /` and `GET /providers` stay public.
- `usage/start|update|stop` and `plan/change|subscribe` name the user in the body, so they now need a call signed by `core/http.Client.Internal` (`HMAC_KEY`) or a platform admin.
- `account`, `payment`, `plan`, `usage`, `voucher` (the caller's workspace) and `transaction` (the caller's own) are readable by a signed-in caller and writable only by a platform admin; `rate` and `plan-type` likewise, unscoped. `payment-log` and `voucher-type` (it holds redeem tokens) are platform-admin only. Before, any signed-in caller could read and write every tenant's rows.
- Saved cards (`provider`) are scoped to accounts in the caller's workspace; the owner may change only `primary`, and only an admin creates one by hand.
- `verify/:type` completes only a payment of the caller's own plan: by reference alone it linked another person's card to the caller's account. `pay` / `charge` with a `providerId` use only a card on the plan owner's own account.
- The flags app is imported for its services only: its `flag` CRUD was served, unguarded, as `/apps/capital/flag`. That path is gone.
- The bus handlers (off unless a host registers `Subscriptions`) are fixed: `apps.capital.payment.debt` saved the owed total as one more unpaid row, so every debt run doubled the debt (it records nothing now); `payment.pay` no longer records again a payment PaymentService has saved; usage events from `UsageService` were ignored (no `rateId`) and are recorded; heartbeat accruals happen once per period; sign-up's two wallet events make one account.

## common v0.0.10 — tag and project links stay in the caller's workspace

Security fix (PLAN U26). **Breaking** for anyone who relied on the open routes.

- `tag-type` and `project-type` `sync`, `attach` and `detach` need a session (before, anonymous callers could link and unlink tags and projects on any entity) and work only inside the caller's own user, client and workspace: new links take them from the session, sync reads only the caller's links, an id resolves only to the caller's own type, and a caller with no workspace gets 403.
- `project`, `project-type`, `tag` and `tag-type` are scoped to the caller's workspace (list, read, write). `rate` is readable by a signed-in caller and writable by a platform admin; `ip` and `rate-log` are platform-admin only. `currency/:currency` and `location` stay public.

## identity v0.0.17 — sign out one session with its refresh token

Additive. `POST apps/identity/logout` accepts `{"refreshToken": "..."}` and ends that one session by deleting its record, so a native app can sign out of one phone without signing out the others or leaving the token usable. The token must verify; signing out twice is a 200. Without `refreshToken` the route behaves as before.

## identity v0.0.16 — `POST /refresh` for native apps

Additive; nothing existing changes.

Sign-in has always returned a refresh token, but no route accepted it, so a native app could not stay signed in once the access token expired. `POST apps/identity/refresh` with `{"token": "<refreshToken>"}` returns `{accessToken, user}`.

- Refused with one 401, whatever the reason: not a refresh token on record (signed out), expired, issued before a "sign out everywhere", account closed or suspended, or a workspace the person has left.
- Claims are rebuilt from the database, so a role removed since sign-in is not carried into the new token.
- Issues an access token only and writes nothing; the refresh token is not rotated.
- The browser shell is unaffected: it keeps re-minting through its device cookie (`GET /sessions`).

## identity v0.0.15 — RBAC writes are authorised

Use v0.0.15: v0.0.14 carries the same fix but does not build outside the workspace.

Security fix. Anyone could grant themselves any role: `POST /role/attach`, `detach`, `sync` and the `permission` equivalents needed no session, and `ResolveRoles` ignored the user, so a role attached to one person reached their whole workspace.

- `role` and `permission` link routes need a session.
- Writes to `role`, `role-type`, `permission` and `permission-type` need a platform admin (`ROOT_USER_ID`, or a `root` / `global-admin` role) or the owner or admin of the token's workspace. Ordinary members get 403.
- A workspace admin's role links stay in their own client and workspace, and can't touch `root`, `global-admin` or `global-member`.
- A role attached to a user now applies to that user only. Rows with no user entity still apply to everyone in scope.
- Tokens add the name of each attached role type to `roles`, beside the existing id.

**Opt-in:** `IDENTITY_RBAC_STRICT=true` makes every RBAC write platform-admin only. Role types and permissions are global tables, so by default a workspace owner can still change them for every tenant; the console's Roles and Permissions screens depend on that. Self-serve products should turn this on.

**Check your data:** any `IdentityRoles` row you can't account for, especially one naming `global-admin`, may have been created through the open routes.

## [Unreleased]

### Added

- **audit** — `pkg/batch/SummaryBatch`: hourly background aggregation of audit log metrics, triggered by the `apps.cron.heartbeat.hourly` event subscription
- **common** — `GetLocation` now calls [ip-api.com](http://ip-api.com) for real geo-IP resolution (country + city) and falls back to storing `US` when the lookup fails; configurable via `GEO_API_URL`
- **capital** — `wallet` resource: `POST /wallet/init`, `GET /wallet/detail`, `GET /wallet/balance`, `POST /wallet/upgrade`, `GET /wallet/transactions`
- **figs** — `/file` CRUD REST resource exposed via router
- **figs** — Real Excel converter using `xuri/excelize/v2`
- **figs** — Real PDF converter using `go-pdf/fpdf`
- **figs** — JSON Schema validation using `xeipuuv/gojsonschema`
- **figs** — YAML Schema validation (parses YAML, validates against JSON Schema)
- **figs** — S3-compatible object store using `aws-sdk-go-v2`
- **health** — `pkg/task/SummaryTask`: 10-minute background health summary computation
- **identity** — Full `AuthService` with Bearer, Basic, API Token, and HMAC auth methods; role and permission resolution embedded in JWT claims
- **notification** — `pkg/provider/ProviderService`: Mustache template rendering, Mailgun email, Zoho Mail email, web (Message record) delivery
- **polylog** — `pkg/pipeline/PipelineService`: `POST /ingest` endpoint, async event processing, `WebhookSink` dispatch

### Changed

- **blobs** (gox-packages) — `S3Provider` is now fully implemented using `aws-sdk-go-v2`; `BlobsConfig` gains `S3AccessKey` / `S3SecretKey` fields

### Added (gox-packages)

- **core** — `image` package: SVG generator with solid, gradient, and pixel variants

---

## Previous

See [VERIFY.md](VERIFY.md) for the full feature-parity audit between gox-apps and ntx-apps.
