// Package auth ports the email/password auth surface from
// ntx-apps/libs/identity/src/app.controller.ts (there is no separate TS
// auth/identity controller — all of /initiate, /verify, /secret, /login,
// /reset-secret/{initiate,verify,update} and /attributes live on that one
// file, :101 onward) into gox's own existing schema (User, Attribute, UCW,
// Workspace, Device, DeviceSession, DeviceLog, RoleType, PermissionType,
// Permission — the same entities identity/app/provider's AuthorizeController
// already uses for the OAuth login flow).
//
// This is a scoped port, not a byte-for-byte one — PLAN M0-27a needs guard's
// sign-up/verify/sign-in/reset-password flows working end to end against
// Origine's actual schema, not TS's. Deliberately left out (see the research
// this was ported from for the full TS behaviour):
//
//   - The invite flow (workspaceId/roleTypeId threaded through initiate →
//     verify → secret). Origine's Invite entity is missing the ref/type/
//     stage/roleTypeId columns TS's version filters and updates on, and
//     guard's own initiate/verify/secret forms never send these fields
//     today. Building it out is a separate, additive-migration task.
//   - AES-256-CBC envelope encryption around the bcrypt hash. TS encrypts
//     the bcrypt digest at rest (ENCRYPTION_KEY) to interoperate with an
//     existing TS-created Users table; Origine's User table has no such
//     legacy data to stay compatible with, so the secret column just holds
//     the bcrypt hash directly (security.Hash/Compare, gox-packages/core).
//   - Per-user Role *assignment* (a Role row linking a user to a RoleType).
//     seedDefaultRoles below only ensures the global RoleType/PermissionType
//     /Permission catalog rows exist, exactly like provider/authorize
//     .controller.go's already-shipped seedDefaultRoles does for the OAuth
//     path — neither path assigns a user a Role row today. Real per-user
//     RBAC assignment is a bigger, pre-existing gap this task didn't create
//     and isn't scoped to fix.
//   - TS's per-workspace token minting and "platform role" computation on
//     /login. Token issuance here reuses identity/pkg's already-live
//     AuthService.Login/IssueTokensByHandle (the same one
//     provider/authorize.controller.go already calls after an OAuth
//     exchange) for one access+refresh pair, rather than inventing a third,
//     divergent token shape.
//   - Cookie-based device continuity (`x_ntx_device_id`). gox's
//     response.Success has no header/cookie parameter, so each login
//     bootstraps a fresh Device+DeviceSession pair instead of reusing one
//     via a cookie the client sends back.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/awesome-goose/goose/types"
	identityattribute "github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	identitydevice "github.com/thescaffold/gox-apps/libs/identity/app/device"
	identitydevicelog "github.com/thescaffold/gox-apps/libs/identity/app/devicelog"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identitypermissiontype "github.com/thescaffold/gox-apps/libs/identity/app/permissiontype"
	identityroletype "github.com/thescaffold/gox-apps/libs/identity/app/roletype"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	identitypkg "github.com/thescaffold/gox-apps/libs/identity/pkg"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/security"
	"github.com/thescaffold/gox-packages/libs/core/services"
	"github.com/thescaffold/gox-packages/libs/core/services/otp"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// otpToken is the decoded shape of the base64 `token` field every
// verify/secret/reset-secret handler receives — the otp.Service's own
// `OTP.Hashed` field, round-tripped. Field order matches otp.OTP's JSON tags
// (value, owner, type, userId), which in turn mirrors TS's object-literal
// insertion order.
type otpToken struct {
	Value  string `json:"value"`
	Owner  string `json:"owner"`
	Type   string `json:"type"`
	UserID string `json:"userId"`
}

func decodeToken(token string) (otpToken, error) {
	raw, err := security.FromBase64(token)
	if err != nil {
		return otpToken{}, err
	}
	var out otpToken
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return otpToken{}, err
	}
	if out.Value == "" || out.Owner == "" || out.UserID == "" {
		return otpToken{}, errors.New("malformed token")
	}
	return out, nil
}

// registerInitiateEntry / resetInitiateEntry are the cache payloads Initiate
// and ResetSecretInitiate stash, mirroring TS's
// `user:register:initiate:<ref>:<type>` / `user:secret-reset:initiate:<ref>:<type>`
// cache entries (minus the invite-only sessionId field).
type registerInitiateEntry struct {
	Id       string `json:"id"`
	Ref      string `json:"ref"`
	Type     string `json:"type"`
	ClientId string `json:"clientId"`
}

const (
	registerInitiatePrefix = "identity:register:initiate:"
	registerVerifyPrefix   = "identity:register:verify:"
	resetInitiatePrefix    = "identity:reset:initiate:"
	tempEntryTTL           = 24 * time.Hour
)

// AuthController hosts the email/password auth surface — see the package
// doc for what this does and doesn't port from TS.
type AuthController struct {
	userEntity           *identityuser.UserEntity                     `inject:""`
	attributeEntity      *identityattribute.AttributeEntity           `inject:""`
	workspaceEntity      *identityworkspace.WorkspaceEntity           `inject:""`
	ucwEntity            *identityucw.UserClientWorkspaceEntity       `inject:""`
	deviceEntity         *identitydevice.DeviceEntity                 `inject:""`
	deviceSessionEntity  *identitydevicesession.DeviceSessionEntity   `inject:""`
	deviceLogEntity      *identitydevicelog.DeviceLogEntity           `inject:""`
	roleTypeEntity       *identityroletype.RoleTypeEntity             `inject:""`
	permissionTypeEntity *identitypermissiontype.PermissionTypeEntity `inject:""`
	permissionEntity     *identitypermission.PermissionEntity         `inject:""`
	cache                services.CacheBackend                        `inject:""`
	tracker              *events.TrackerService                       `inject:""`
	lang                 *i18n.Service                                `inject:""`
	auth                 *identitypkg.AuthService                     `inject:""`

	// otpSvc is constructed lazily from the injected CacheBackend (Boot,
	// below) rather than trusting DI to inject *otp.Service directly — the
	// upstream research for this task flagged that goose's container may not
	// register a CoreModule-declared root instance as a resolvable binding,
	// which would hand this a zero-value Service (nil cache). Constructing
	// it ourselves from a field we already know is wired removes the doubt.
	otpSvc *otp.Service
}

func (c *AuthController) Boot(_ types.Kernel) error {
	if c.otpSvc == nil {
		c.otpSvc = otp.New(c.cache, 7*time.Minute)
	}
	return nil
}

// Initiate mirrors TS POST /initiate (app.controller.ts:697-764) — register
// step 1: reject a duplicate ref, otherwise mint (or resend) an OTP under a
// temporary id and cache the pending registration for 24h.
func (c *AuthController) Initiate(dto *InitiateDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	if existing, _ := c.userEntity.First(`"ref" = ? OR "email" = ?`, dto.Ref, dto.Ref); existing != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.initiate.error.duplicate", nil, pref))
	}

	key := registerInitiatePrefix + dto.Ref + ":" + dto.Type
	id := ""
	if cached, ok := c.cache.Get(key); ok {
		var entry registerInitiateEntry
		if json.Unmarshal([]byte(cached), &entry) == nil {
			id = entry.Id
		}
	}
	if id == "" {
		id = utils.UUID()
		entry := registerInitiateEntry{Id: id, Ref: dto.Ref, Type: dto.Type, ClientId: dto.ClientId}
		bytes, _ := json.Marshal(entry)
		c.cache.Set(key, string(bytes), tempEntryTTL)
	}

	rec := c.otpSvc.Initiate(dto.Type, id, dto.Ref, "register")
	c.sendVerificationEmail(dto.Ref, id, dto.ClientId, rec.Value)

	return response.Success(map[string]any{"id": id}, title,
		c.lang.Translate("apps.identity.app.post.initiate.success", map[string]any{"type": dto.Type}, pref), nil)
}

// Verify mirrors TS POST /verify (app.controller.ts:766-919) — register step
// 2. Verifies the OTP without clearing it (the same token is re-verified,
// clearing it, by Secret) and stashes the pending registration under its
// permanent-looking `id` for Secret to pick up.
func (c *AuthController) Verify(dto *VerifyDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	tok, err := decodeToken(dto.Token)
	if err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.verify.error.expired", nil, pref))
	}

	key := registerInitiatePrefix + tok.Owner + ":" + tok.Type
	if _, ok := c.cache.Get(key); !ok {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.verify.error.expired", nil, pref))
	}

	if _, err := c.otpSvc.Verify(tok.Value, tok.Type, tok.UserID, tok.Owner, "register", false); err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.verify.error.mismatch", nil, pref))
	}

	entry := registerInitiateEntry{Id: tok.UserID, Ref: tok.Owner, Type: tok.Type}
	if cached, ok := c.cache.Get(key); ok {
		_ = json.Unmarshal([]byte(cached), &entry)
	}
	bytes, _ := json.Marshal(entry)
	c.cache.Set(registerVerifyPrefix+tok.UserID, string(bytes), tempEntryTTL)

	return response.Success(map[string]any{"id": tok.UserID, "userAlreadyExist": false}, title,
		c.lang.Translate("apps.identity.app.post.verify.success.new", nil, pref), nil)
}

// Secret mirrors TS POST /secret (app.controller.ts:921-1194) — register
// step 3: set the password, create the account, bootstrap a workspace, and
// sign the user in.
func (c *AuthController) Secret(dto *SecretDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	if dto.Secret != dto.ConfirmSecret {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.secret.error.mismatch", nil, pref))
	}

	tok, err := decodeToken(dto.Token)
	if err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.secret.error.unknown", nil, pref))
	}
	if _, err := c.otpSvc.Verify(tok.Value, tok.Type, tok.UserID, tok.Owner, "register", true); err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.secret.error.unknown", nil, pref))
	}

	verifyKey := registerVerifyPrefix + tok.UserID
	entry := registerInitiateEntry{Id: tok.UserID, Ref: tok.Owner, Type: tok.Type}
	if cached, ok := c.cache.Get(verifyKey); ok {
		_ = json.Unmarshal([]byte(cached), &entry)
	} else {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.secret.error.unknown", nil, pref))
	}

	hashed, err := security.Hash(dto.Secret)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	name := strings.TrimSpace(dto.FirstName + " " + dto.LastName)
	if name == "" {
		name = entry.Ref
	}
	ref, userType := entry.Ref, entry.Type
	newUser := &identityuser.User{
		Ref: &ref, Type: &userType, Name: name, Email: entry.Ref, Secret: &hashed,
	}
	if err := c.userEntity.Insert(newUser); err != nil {
		return response.InternalServerError(title, err.Error())
	}

	ws := c.bootstrapWorkspace(newUser, entry.ClientId)

	if dto.ReferralCode != "" {
		c.redeemReferralCode(newUser.Id, dto.ReferralCode)
	}
	c.upsertAttributes(newUser.Id, "user", map[string]string{
		"first_name": dto.FirstName,
		"last_name":  dto.LastName,
		"email":      entry.Ref,
		"username":   strings.ToLower(strings.ReplaceAll(name, " ", "-")),
	})
	c.upsertAttributes(newUser.Id, "preference", map[string]string{
		"theme": dto.Theme, "language": dto.Language, "timezone": dto.Timezone,
		"country": dto.Country, "currency": dto.Currency,
	})
	c.seedDefaultRoles()
	c.cache.Del(verifyKey)

	access, refresh, err := c.auth.Login(entry.Ref, dto.Secret, entry.ClientId)
	if err != nil {
		return response.Unauthorized(title, err.Error())
	}

	var workspaces []identityworkspace.Workspace
	if ws != nil {
		workspaces = []identityworkspace.Workspace{*ws}
	}
	return response.Success(map[string]any{
		"user":         maskUser(newUser),
		"workspaces":   workspaces,
		"accessToken":  access,
		"refreshToken": refresh,
	}, title, c.lang.Translate("apps.identity.app.post.secret.success", nil, pref), nil)
}

// Login mirrors TS POST /login (app.controller.ts:1196-1363). Token issuance
// and credential verification are identity/pkg's already-live
// AuthService.Login; this handler layers on the workspace/attribute/device
// bootstrap guard's gate.component.ts actually consumes.
func (c *AuthController) Login(dto *LoginDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	access, refresh, err := c.auth.Login(dto.Ref, dto.Secret, dto.ClientId)
	if err != nil {
		return response.Unauthorized(title,
			c.lang.Translate("apps.identity.app.post.login.error.invalid-request", nil, pref))
	}
	u, _ := c.userEntity.First(`"ref" = ? OR "email" = ?`, dto.Ref, dto.Ref)
	if u == nil {
		return response.Unauthorized(title,
			c.lang.Translate("apps.identity.app.post.login.error.invalid-request", nil, pref))
	}

	if closed, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "closed_at"); closed != nil {
		return response.Unauthorized(title,
			c.lang.Translate("apps.identity.app.post.login.error.invalid-request", nil, pref))
	}
	if suspended, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "suspended_at"); suspended != nil {
		return response.Unauthorized(title,
			c.lang.Translate("apps.identity.app.post.login.error.suspended", nil, pref))
	}

	workspaces := c.workspacesForUser(u.Id, dto.ClientId)
	attributes := c.attributesForUser(u.Id)
	if len(workspaces) > 0 {
		deviceId, sessionId := c.bootstrapDevice(u.Id, dto.UserAgent)
		c.logDeviceLogin(deviceId, sessionId, dto.Ctx)
	}

	return response.Success(map[string]any{
		"user":         maskUser(u),
		"accessToken":  access,
		"refreshToken": refresh,
		"workspaces":   workspaces,
		"attributes":   attributes,
	}, title, c.lang.Translate("apps.identity.app.post.login.success", nil, pref), nil)
}

// ResetSecretInitiate mirrors TS POST /reset-secret/initiate
// (app.controller.ts:1988-2033). Always returns success regardless of
// whether ref matches an account, so the response can't be used to enumerate
// accounts.
func (c *AuthController) ResetSecretInitiate(dto *InitiateDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	if u, _ := c.userEntity.First(`"ref" = ? OR "email" = ?`, dto.Ref, dto.Ref); u != nil {
		key := resetInitiatePrefix + dto.Ref + ":" + dto.Type
		entry := registerInitiateEntry{Id: u.Id, Ref: dto.Ref, Type: dto.Type, ClientId: dto.ClientId}
		bytes, _ := json.Marshal(entry)
		c.cache.Set(key, string(bytes), tempEntryTTL)

		rec := c.otpSvc.Initiate(dto.Type, u.Id, dto.Ref, "reset")
		c.sendVerificationEmail(dto.Ref, u.Id, dto.ClientId, rec.Value)
	}

	return response.Success(map[string]any{"ref": dto.Ref}, title,
		c.lang.Translate("apps.identity.app.post.reset-secret.initiate.success", nil, pref), nil)
}

// ResetSecretVerify mirrors TS POST /reset-secret/verify (:2035-2074).
func (c *AuthController) ResetSecretVerify(dto *VerifyDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	tok, err := decodeToken(dto.Token)
	if err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.reset-secret.verify.error.expired", nil, pref))
	}
	key := resetInitiatePrefix + tok.Owner + ":" + tok.Type
	if _, ok := c.cache.Get(key); !ok {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.reset-secret.verify.error.expired", nil, pref))
	}
	if _, err := c.otpSvc.Verify(tok.Value, tok.Type, tok.UserID, tok.Owner, "reset", false); err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.reset-secret.verify.error.mismatch", nil, pref))
	}

	return response.Success(map[string]any{"id": tok.UserID}, title,
		c.lang.Translate("apps.identity.app.post.reset-secret.verify.success", nil, pref), nil)
}

// ResetSecretUpdate mirrors TS POST /reset-secret/update (:2076-2118).
func (c *AuthController) ResetSecretUpdate(dto *ResetSecretUpdateDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	if dto.NewSecret != dto.ConfirmNewSecret {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.reset-secret.update.error.mismatch", nil, pref))
	}
	tok, err := decodeToken(dto.Token)
	if err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.reset-secret.update.error.mismatch", nil, pref))
	}
	if _, err := c.otpSvc.Verify(tok.Value, tok.Type, tok.UserID, tok.Owner, "reset", true); err != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.app.post.reset-secret.update.error.mismatch", nil, pref))
	}

	hashed, err := security.Hash(dto.NewSecret)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	if _, err := c.userEntity.Update(&identityuser.User{Secret: &hashed}, `"id" = ?`, tok.UserID); err != nil {
		return response.InternalServerError(title, err.Error())
	}

	return response.Success(map[string]any{"id": tok.UserID, "ref": tok.Owner}, title,
		c.lang.Translate("apps.identity.app.post.reset-secret.update.success", nil, pref), nil)
}

// Attributes mirrors TS GET /attributes (:1563-1616) — always scoped to the
// current authenticated user.
func (c *AuthController) Attributes(dto *AttributesQueryDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	if dto.Ctx.UserID == "" {
		return response.Unauthorized(title,
			c.lang.Translate("apps.identity.app.get.attributes.error.unauthorized", nil, pref))
	}
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title,
			c.lang.Translate("apps.identity.app.get.attributes.error.unauthorized", nil, pref))
	}

	attributes := c.attributesForUserFiltered(dto.Ctx.UserID, dto.Type)
	workspaces := c.workspacesForUser(dto.Ctx.UserID, dto.Ctx.ClientID)

	out := maskUser(u)
	out["attributes"] = attributes
	out["workspaces"] = workspaces
	return response.Success(out, title,
		c.lang.Translate("apps.identity.app.get.attributes.success", nil, pref), nil)
}

// createOnlyAttributeKeys mirrors TS CreateOnlyUserAttributeType
// (app.config.ts:17-21) — set-once preference fields a later PATCH can't
// silently overwrite.
var createOnlyAttributeKeys = map[string]bool{"timezone": true, "country": true, "currency": true}

// UpdateAttribute mirrors TS PATCH /attributes/:type (:1618-1711).
func (c *AuthController) UpdateAttribute(dto *UpdateAttributesDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	if dto.Ctx.UserID == "" {
		return response.Unauthorized(title,
			c.lang.Translate("apps.identity.app.patch.attributes.error.unauthorized", nil, pref))
	}
	if dto.Type == "" {
		return response.NotFound(title,
			c.lang.Translate("apps.identity.app.patch.attributes.error.not-found", nil, pref))
	}

	for key, raw := range dto.Body {
		value := fmt.Sprintf("%v", raw)
		if createOnlyAttributeKeys[key] {
			if existing, _ := c.attributeEntity.First(
				`"user_id" = ? AND "type" = ? AND "key" = ?`, dto.Ctx.UserID, dto.Type, key,
			); existing != nil {
				continue
			}
		}
		c.upsertAttribute(dto.Ctx.UserID, dto.Type, key, value)
	}

	if c.tracker != nil {
		c.tracker.Message("apps.identity.attribute.type.update", map[string]any{
			"user": dto.Ctx.UserID, "type": dto.Type,
		})
	}

	return response.Success(map[string]any{"attributes": c.attributesForUser(dto.Ctx.UserID)}, title,
		c.lang.Translate("apps.identity.app.patch.attributes.success", nil, pref), nil)
}

// --- helpers ---

// sendVerificationEmail dispatches the OTP over email, best-effort — a nil
// tracker (or one that no notification channel is wired to) just means no
// email goes out; it never fails the request.
func (c *AuthController) sendVerificationEmail(ref, userId, clientId, code string) {
	if c.tracker == nil {
		return
	}
	c.tracker.Message("apps.notification.message.new", map[string]any{
		"reference": utils.Reference("OTP", 32),
		"userId":    userId,
		"clientId":  clientId,
		"key":       "email-verification",
		"channels":  []string{"email"},
		"data":      map[string]any{"code": code, "ref": ref},
		"subject":   "Email Verification",
		"type":      "security",
		"priority":  "medium",
		"scope":     "user",
	})
}

// bootstrapWorkspace creates a default workspace + owner UCW row for a new
// user. Same shape as provider/authorize.controller.go's bootstrapWorkspace.
func (c *AuthController) bootstrapWorkspace(u *identityuser.User, clientId string) *identityworkspace.Workspace {
	if c.workspaceEntity == nil || c.ucwEntity == nil {
		return nil
	}
	ws := &identityworkspace.Workspace{
		UserId: u.Id, ClientId: clientId, Name: "My Workspace", Slug: utils.Reference("WS", 16),
	}
	if err := c.workspaceEntity.Insert(ws); err != nil {
		return nil
	}
	owner := "owner"
	_ = c.ucwEntity.Insert(&identityucw.UserClientWorkspace{
		UserId: u.Id, ClientId: clientId, WorkspaceId: ws.Id, Status: &owner,
	})
	return ws
}

// workspacesForUser mirrors provider/authorize.controller.go's helper of the
// same name.
func (c *AuthController) workspacesForUser(userId, clientId string) []identityworkspace.Workspace {
	if c.ucwEntity == nil || c.workspaceEntity == nil {
		return nil
	}
	links, _ := c.ucwEntity.Find(0, 0, `"user_id" = ? AND "client_id" = ?`, userId, clientId)
	out := make([]identityworkspace.Workspace, 0, len(links))
	for _, link := range links {
		if w, err := c.workspaceEntity.First(`"id" = ?`, link.WorkspaceId); err == nil && w != nil {
			out = append(out, *w)
		}
	}
	return out
}

// attributesForUser groups every Attribute row for userId into
// {[type]: {[key]: value}}, mirroring TS attributeService.getAll.
func (c *AuthController) attributesForUser(userId string) map[string]map[string]string {
	return c.attributesForUserFiltered(userId, "")
}

func (c *AuthController) attributesForUserFiltered(userId, attrType string) map[string]map[string]string {
	out := map[string]map[string]string{}
	if c.attributeEntity == nil {
		return out
	}
	var rows []identityattribute.Attribute
	if attrType != "" {
		rows, _ = c.attributeEntity.Find(0, 0, `"user_id" = ? AND "type" = ?`, userId, attrType)
	} else {
		rows, _ = c.attributeEntity.Find(0, 0, `"user_id" = ?`, userId)
	}
	for _, row := range rows {
		if row.Type == nil || row.Value == nil {
			continue
		}
		if out[*row.Type] == nil {
			out[*row.Type] = map[string]string{}
		}
		out[*row.Type][row.Key] = *row.Value
	}
	return out
}

// upsertAttribute writes one (type, key) => value row, idempotently.
func (c *AuthController) upsertAttribute(userId, attrType, key, value string) {
	if c.attributeEntity == nil {
		return
	}
	existing, _ := c.attributeEntity.First(`"user_id" = ? AND "type" = ? AND "key" = ?`, userId, attrType, key)
	if existing != nil {
		v := value
		existing.Value = &v
		_, _ = c.attributeEntity.Update(existing, `"id" = ?`, existing.Id)
		return
	}
	t := attrType
	v := value
	_ = c.attributeEntity.Insert(&identityattribute.Attribute{UserId: userId, Key: key, Value: &v, Type: &t})
}

// upsertAttributes upserts several (type, key) => value pairs, skipping
// blank values (mirrors the "silently drop empty fields" behaviour used
// throughout the ported handlers above).
func (c *AuthController) upsertAttributes(userId, attrType string, kv map[string]string) {
	for key, value := range kv {
		if value == "" {
			continue
		}
		c.upsertAttribute(userId, attrType, key, value)
	}
}

// redeemReferralCode mirrors provider/authorize.controller.go's helper —
// duplicated rather than shared because that one writes the camelCase
// `referralCode`/`referrerId` keys the OAuth path already established, while
// this one (freshly porting from TS, which was never ported into the OAuth
// path) uses TS's own snake_case `referral_code`/`referrer_id`. The two
// paths' attribute keys are already inconsistent with each other before this
// change; reconciling that is a separate cleanup, not this task's job.
func (c *AuthController) redeemReferralCode(userId, code string) {
	if c.attributeEntity == nil {
		return
	}
	referrer, _ := c.attributeEntity.First(`"type" = ? AND "key" = ? AND "value" = ?`, "user", "referral_code", code)
	if referrer == nil {
		return
	}
	c.upsertAttribute(userId, "user", "referrer_id", referrer.UserId)
}

// bootstrapDevice creates a fresh Device + DeviceSession pair (see the
// package doc for why this doesn't reuse a cookie-carried device id).
func (c *AuthController) bootstrapDevice(userId, userAgent string) (string, string) {
	if c.deviceEntity == nil || c.deviceSessionEntity == nil {
		return "", ""
	}
	dev := &identitydevice.Device{
		UserId: userId, Fingerprint: utils.Reference("DEV", 24), Type: "browser",
	}
	if userAgent != "" {
		dev.Agent = &userAgent
	}
	if err := c.deviceEntity.Insert(dev); err != nil {
		return "", ""
	}
	exp := time.Now().UTC().Add(24 * time.Hour)
	sess := &identitydevicesession.DeviceSession{
		DeviceId: dev.Id, UserId: userId, Token: utils.Reference("SES", 32), ExpiresAt: &exp,
	}
	if err := c.deviceSessionEntity.Insert(sess); err != nil {
		return dev.Id, ""
	}
	return dev.Id, sess.Id
}

// logDeviceLogin records a DeviceLog row for a successful login. Same shape
// as provider/authorize.controller.go's helper, minus the OAuth `profile`
// object this path never has.
func (c *AuthController) logDeviceLogin(deviceId, sessionId string, ntx ntxctx.NTXContext) {
	if c.deviceLogEntity == nil || deviceId == "" {
		return
	}
	country, _ := ntx.Preference["country"].(string)
	meta, _ := json.Marshal(map[string]any{
		"theme": ntx.Preference["theme"], "language": ntx.Preference["language"],
		"timezone": ntx.Preference["timezone"], "currency": ntx.Preference["currency"],
	})
	_ = c.deviceLogEntity.Insert(&identitydevicelog.DeviceLog{
		DeviceId: deviceId, SessionId: &sessionId, Event: "login", Country: &country, Meta: meta,
	})
}

// seedDefaultRoles ensures the global Admin/Member RoleType + wildcard
// PermissionType + Permission catalog rows exist. Identical in shape to
// provider/authorize.controller.go's helper of the same name — see the
// package doc's note on per-user Role assignment not being done by either
// path.
func (c *AuthController) seedDefaultRoles() {
	if c.roleTypeEntity == nil || c.permissionTypeEntity == nil || c.permissionEntity == nil {
		return
	}
	defaults := []struct{ roleName, permName, permVerb, permScope string }{
		{"Admin", "admin:*:*:*:*:*:workspace", "*", "workspace"},
		{"Member", "member:*:*:*:*:workspace", "update", "workspace"},
	}
	for _, def := range defaults {
		rt, _ := c.roleTypeEntity.First(`"name" = ?`, def.roleName)
		if rt == nil {
			rt = &identityroletype.RoleType{Name: def.roleName}
			_ = c.roleTypeEntity.Insert(rt)
		}
		pt, _ := c.permissionTypeEntity.First(`"name" = ?`, def.permName)
		if pt == nil {
			pt = &identitypermissiontype.PermissionType{Name: def.permName}
			_ = c.permissionTypeEntity.Insert(pt)
		}
		if existing, _ := c.permissionEntity.First(
			`"role_type_id" = ? AND "permission_type_id" = ?`, rt.Id, pt.Id,
		); existing != nil {
			continue
		}
		rtId, ptId := rt.Id, pt.Id
		_ = c.permissionEntity.Insert(&identitypermission.Permission{
			RoleTypeId: &rtId, PermissionTypeId: &ptId,
			Key: def.permName, Action: def.permVerb, Scope: def.permScope,
		})
	}
}

// maskUser mirrors TS maskUser() / provider/authorize.controller.go's helper
// of the same name — strips Secret before the user envelope reaches a
// client.
func maskUser(u *identityuser.User) map[string]any {
	if u == nil {
		return nil
	}
	out := map[string]any{"id": u.Id, "name": u.Name, "email": u.Email}
	if u.Ref != nil {
		out["ref"] = *u.Ref
	}
	if u.Type != nil {
		out["type"] = *u.Type
	}
	return out
}
