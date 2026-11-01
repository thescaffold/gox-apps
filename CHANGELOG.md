# Changelog

All notable changes to gox-apps are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## identity v0.0.14 — RBAC writes are authorised

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
