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
import "github.com/thescaffold/gox-apps/libs/identity/app"
import authpkg "github.com/thescaffold/gox-apps/libs/identity/pkg"

// Full app
app.AppModule{}

// Auth-only (for services that only need the auth endpoints)
authpkg.AuthModule{}
```

## Development

```bash
go build ./...
# the integration tests need a scratch Postgres database (they skip without one)
IDENTITY_TEST_DB_URL="postgres://postgres@127.0.0.1:5432/identity_test?sslmode=disable" go test ./...
```

## RBAC writes

Role and permission writes need a platform admin: the role-type, permission-type and permission tables are global, so a workspace owner who wrote them would reach every tenant. A console that still edits roles as a workspace owner or admin (never platform roles) sets `IDENTITY_RBAC_STRICT=false`. See `pkg/rbac`.

## API keys belong to their holder

`/keys` works on the keys it issues (those with `IDENTITY_KEY_BRAND`). A person lists, revokes and rotates their own. Owners and admins (`org.tokens.manage`) also list and revoke other people's keys in the workspace, never rotate them, so nobody sees a secret of someone else's key. Keys a host issues under another brand (its machine credentials) are its own to manage.

## Account

`PATCH /account {name}` renames the signed-in person. `POST /account/email {email}` sends a code to the new address; `POST /account/email/verify {email, code}` changes the sign-in email and tells the old address (templates `email-change` and `email-changed`). The time zone stays editable through `PATCH /attributes/preference`; country and currency are set once.

Sign-in, sign-up and passwordless routes refuse a missing `clientId` (400), and a session without one cannot create a workspace. An owner switches into their workspace from any client: the owner membership is recorded for that client.

## Host hooks (`pkg/org`)

- `SetAuthorizer(func(ctx, ws, userID, action) (bool, error))` answers identity's org-level checks: `org.members.manage` (seats, invitations, teams), `org.security.manage` (SSO connection, 2FA policy), `org.settings.manage` (workspace name and handle), `org.tokens.manage` (other people's keys), `org.roles.manage` (custom roles) and `org.audit.read` (activity). Without one, each is "an owner or admin", as before. The seat rules (nobody above their own seat, only an owner grants owner, one owner stays) apply whatever it answers.
- `OnSeatRemoved(func(ctx, tx, ws, userID, actorID) error)` runs inside the transaction that removes a seat, so a host can end what the seat gave (project memberships); an error keeps the seat.
