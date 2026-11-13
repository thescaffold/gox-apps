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

## API keys expire (`IDENTITY_KEY_EXPIRY_REQUIRED`)

With `IDENTITY_KEY_EXPIRY_REQUIRED=true`, `POST /keys` refuses a key with no `expires_in_seconds`, or one past the workspace's maximum, with 422 (`expiry_required`, `expiry_too_long`; `details.maxDays`). The maximum is a year unless an owner or admin (`org.security.manage`) sets a shorter one with `PUT /keys/policy {maxDays}` (1 to 365); `GET /keys/policy` reads it. A rotated key lives as long as the one it replaces, within the maximum. A host issuing keys itself sets `apikey.Service.MaxTTL` (`apikey.WorkspaceMaxTTL(db)`). Without the setting nothing changes.

## Account

`PATCH /account {name}` renames the signed-in person. `POST /account/email {email}` sends a code to the new address; `POST /account/email/verify {email, code}` changes the sign-in email and tells the old address (templates `email-change` and `email-changed`). The time zone stays editable through `PATCH /attributes/preference`; country and currency are set once.

Sign-in, sign-up and passwordless routes refuse a missing `clientId` (400), and a session without one cannot create a workspace. An owner switches into their workspace from any client: the owner membership is recorded for that client.

## Sessions (`/account/sessions`)

A person lists their own device sessions on every device with `GET /account/sessions` (browser and OS from the device's user agent, address, place, last use, `current` for this one) and signs one out with `DELETE /account/sessions/:id`, or every one but this with `POST /account/sessions/revoke-others`. Someone else's session is 404. Every token a sign-in or a device re-mint issues names its session (`sid`), and `revocation.Check` refuses it once the session is signed out, so it ends at once; a refreshed token keeps its `sid`. Sign-in routes record the client's address (`ntxctx.RequestInfo`: set `TRUSTED_PROXIES` behind a proxy); `auth.SetPlaceResolver(func(ip) string)` names its place (without one the list shows the address).

## SAML: signed requests and single logout (`pkg/sso`)

Each workspace's SAML service provider has its own signing key (RSA-2048, a self-signed certificate, sealed with `IDENTITY_TOKEN_KEY`), made the first time an admin opens the connection. `GET /sso/connection` shows its certificate and fingerprint (`spCert`, `spFingerprint`) and the SLO URL (`spSloUrl`); the SP metadata publishes it (signing only: no encryption key, so encrypted assertions stay refused, D-15). `PUT /sso/connection {signRequests}` signs the AuthnRequests (a new connection follows the IdP's `WantAuthnRequestsSigned`). `POST /sso/connection/sp-key {action}` rotates it in two steps: `rotate` stages a next key, published beside the current one; `activate` makes it sign; `cancel` drops it.

Single logout uses the HTTP-Redirect binding, signed both ways, at `GET /sso/saml/:workspace/slo`. Signing out of a browser whose session came through SAML (`POST /logout {device:"yes"}`) answers `logoutUrl`, the IdP's logout address for the browser to go on to; the IdP's answer lands on `SSO_SIGNED_OUT_PATH` (default `/guard/login?signedout=sso`). A LogoutRequest from the IdP (signed by it, for this SLO URL, recent, seen once) ends the sessions it names, by NameID and SessionIndex, and their tokens at once. IdP-initiated sign-in stays refused (D-27).

## Host hooks (`pkg/org`)

- `SetAuthorizer(func(ctx, ws, userID, action) (bool, error))` answers identity's org-level checks: `org.members.manage` (seats, invitations, teams), `org.security.manage` (SSO connection, 2FA policy), `org.settings.manage` (workspace name and handle), `org.tokens.manage` (other people's keys), `org.roles.manage` (custom roles) and `org.audit.read` (activity). Without one, each is "an owner or admin", as before. The seat rules (nobody above their own seat, only an owner grants owner, one owner stays) apply whatever it answers.
- `OnSeatRemoved(func(ctx, tx, ws, userID, actorID) error)` runs inside the transaction that removes a seat, so a host can end what the seat gave (project memberships); an error keeps the seat.
- `SetSeatObserver(func(ctx, tx, SeatChange) error)` sees every seat change identity makes: `SetSeat` and `RemoveSeat` (`Via: "members"`, the owner or admin as actor) and a join through an accepted invitation (`Via: "invitation"`, whoever sent it as actor). `SeatChange` carries the workspace, actor, person, the seats `From` and `To` ("" for none) and, on a removal, the teams left. It runs inside the change's transaction, after `OnSeatRemoved`, so a host's audit row exists if and only if the change does; an error undoes the change. `SeatObserverHook()` (or `ObserveSeat`) lets a host record seat changes it makes itself (SCIM) the same way.
