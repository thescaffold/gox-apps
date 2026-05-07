# Changelog

All notable changes to gox-apps are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

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
