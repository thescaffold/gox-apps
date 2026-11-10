# Changelog

All notable changes to gox-apps are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## capital v0.0.10 — every route says who may call it

Security fix (PLAN U26). **Breaking** for anyone who relied on the open routes; needs core with `coreauth.Caller/Admin/Internal/AdminWrites`.

- The app routes that act on a plan, payment or card (`debt`, `accrual`, `status`, `user-plan`, `init`, `default-provider`, `pay`, `verify/:type`, `:type/reference/:reference`, `:type/charge/:providerId`, `user-plan/compute-upgrade`), the wallet (`wallet/*`) and `voucher/list` / `voucher/verify/:token` need a session and act on the verified caller only. Before, all answered anonymous callers. `GET /` and `GET /providers` stay public.
- `usage/start|update|stop` and `plan/change|subscribe` name the user in the body, so they now need a call signed by `core/http.Client.Internal` (`HMAC_KEY`) or a platform admin.
- `account`, `payment`, `plan`, `usage`, `voucher` (the caller's workspace) and `transaction` (the caller's own) are readable by a signed-in caller and writable only by a platform admin; `rate` and `plan-type` likewise, unscoped. `payment-log` and `voucher-type` (it holds redeem tokens) are platform-admin only. Before, any signed-in caller could read and write every tenant's rows.
- Saved cards (`provider`) are scoped to accounts in the caller's workspace; the owner may change only `primary`, and only an admin creates one by hand.
- `verify/:type` completes only a payment of the caller's own plan: by reference alone it linked another person's card to the caller's account. `pay` / `charge` with a `providerId` use only a card on the plan owner's own account.
- The flags app is imported for its services only: its `flag` CRUD was served, unguarded, as `/apps/capital/flag`. That path is gone.

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
