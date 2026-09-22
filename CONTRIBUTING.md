# Contributing to gox-apps

## Adding a new service

1. Create `libs/<name>/go.mod` with module path `github.com/thescaffold/gox-apps/libs/<name>`.
2. Add the module path to `go.work`.
3. Follow the existing directory layout:
   ```
   libs/<name>/
   ├── app/
   │   ├── app.module.go     # Goose module
   │   ├── app.controller.go
   │   ├── app.service.go
   │   ├── app.routes.go
   │   ├── app.dtos.go
   │   └── index.go          # Name, Subscriptions, DefaultPermissions
   ├── migrations/
   └── tests/
   ```
4. Wire `module.New(module.CoreConfig{})` and your migrations into `AppModule.Imports()`.
5. Add a README at `libs/<name>/README.md`.
6. Add the service to the CI matrix in `.github/workflows/ci.yml`.

## Modifying an existing service

- Keep entity changes paired with a new migration file.
- Do not import across service boundaries — use HTTP calls or the event bus.
- Run `go build ./...` and `go test ./...` from the service directory before opening a PR.

## Publishing a release

Tag releases using the format `libs/<service>/v<semver>` — the nested-module
form Go requires for a module that isn't at the repo root. A bare
`<service>/v<semver>` tag addresses the repo-root module (which doesn't
exist here) instead, and produces a confusing "unknown revision"/module
lookup error unrelated to the actual tag:

```bash
git tag libs/audit/v1.2.0
git push origin libs/audit/v1.2.0
```

The publish workflow will build, test, and create a GitHub Release automatically.

## Code style

- Standard Go formatting (`gofmt`).
- No comments unless the WHY is non-obvious.
- Table-driven tests in `tests/` using `testing.T`.
- Avoid adding dependencies for things the standard library handles.
