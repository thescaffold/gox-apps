# identity

Go/Goose microservice for authentication, authorisation, and user/workspace management.

## Overview

`identity` is the auth backbone of the platform. It manages users, clients, workspaces, roles, permissions, devices, sessions, invites, OAuth providers, and tokens. The `pkg/auth` sub-package provides a full `AuthService` supporting Bearer (JWT), Basic, API Token, and HMAC authentication, plus role/permission resolution.

## Endpoints

### Auth (`pkg/auth`)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/login` | Password login → access + refresh tokens |
| POST | `/auth/logout` | Invalidate refresh token by ID |
| POST | `/auth/refresh` | Exchange refresh token for a new access token |
| GET | `/auth/me` | Return current user claims (requires Bearer token) |
| GET | `/oauth/:provider` | OAuth redirect |
| GET | `/oauth/:provider/callback` | OAuth callback |

### Core CRUD

| Path | Entities |
|------|----------|
| `/attributes` | Attribute |
| `/clients`, `/clientlogs` | Client, ClientLog |
| `/devices`, `/devicelogs`, `/devicesessions` | Device, DeviceLog, DeviceSession |
| `/invites` | Invite |
| `/permissions`, `/permissiontypes` | Permission, PermissionType |
| `/providers`, `/providerlog` | Provider, ProviderLog |
| `/roles`, `/roletypes` | Role, RoleType |
| `/tokens` | Token |
| `/users` | User |
| `/userclientworkspaces` | UserClientWorkspace |
| `/workspaces` | Workspace |

## Auth Methods

| Method | Header | Description |
|--------|--------|-------------|
| Bearer | `Authorization: Bearer <jwt>` | JWT verification |
| Basic | `Authorization: Basic <base64(email:password)>` | Password auth |
| Token | `Authorization: Token <api-token>` | Stored API token |
| HMAC | `Authorization: hmac <signature>` | HMAC signature (for internal service calls) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | `changeme` | JWT signing secret |
| `HMAC_KEY` | `changeme-hmac` | HMAC signing key |

## Module

```go
import "github.com/thescaffold/gox-apps-identity/app"
import authpkg "github.com/thescaffold/gox-apps-identity/pkg"

// Full app
app.AppModule{}

// Auth-only (for services that only need the auth endpoints)
authpkg.AuthModule{}
```

## Development

```bash
go test ./...
go build ./...
```
