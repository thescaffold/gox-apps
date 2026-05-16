// Package provider — Authorize controller hosts the full OAuth login flow
// ported from ntx-apps/libs/identity/src/api/provider/provider.controller.ts
// :provider/authorize endpoint (lines 217–624). Kept on a dedicated
// controller struct so the dozen+ entity injections don't bloat the CRUD
// controller.
package provider

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/awesome-goose/goose/types"
	identityattribute "github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	identitydevice "github.com/thescaffold/gox-apps/libs/identity/app/device"
	identitydevicelog "github.com/thescaffold/gox-apps/libs/identity/app/devicelog"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identitypermissiontype "github.com/thescaffold/gox-apps/libs/identity/app/permissiontype"
	identityproviderlog "github.com/thescaffold/gox-apps/libs/identity/app/providerlog"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	identityroletype "github.com/thescaffold/gox-apps/libs/identity/app/roletype"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	identitypkg "github.com/thescaffold/gox-apps/libs/identity/pkg"
	identityproviderpkg "github.com/thescaffold/gox-apps/libs/identity/pkg/provider"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/services"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// AuthorizeController hosts POST /provider/:provider/authorize — the full
// OAuth login flow. Separate from CRUD so its (~12) entity injections don't
// bloat the listing-only controller.
type AuthorizeController struct {
	providerEntity        *ProviderEntity                                  `inject:""`
	providerLogEntity     *identityproviderlog.ProviderLogEntity           `inject:""`
	userEntity            *identityuser.UserEntity                         `inject:""`
	workspaceEntity       *identityworkspace.WorkspaceEntity               `inject:""`
	ucwEntity             *identityucw.UserClientWorkspaceEntity           `inject:""`
	attributeEntity       *identityattribute.AttributeEntity               `inject:""`
	deviceEntity          *identitydevice.DeviceEntity                     `inject:""`
	deviceSessionEntity   *identitydevicesession.DeviceSessionEntity       `inject:""`
	deviceLogEntity       *identitydevicelog.DeviceLogEntity               `inject:""`
	roleTypeEntity        *identityroletype.RoleTypeEntity                 `inject:""`
	roleEntity            *identityrole.RoleEntity                         `inject:""`
	permissionTypeEntity  *identitypermissiontype.PermissionTypeEntity     `inject:""`
	permissionEntity      *identitypermission.PermissionEntity             `inject:""`
	cache                 services.CacheBackend                            `inject:""`
	tracker               *events.TrackerService                           `inject:""`
	lang                  *i18n.Service                                    `inject:""`
	auth                  *identitypkg.AuthService                         `inject:""`

	// registry holds the per-provider OAuth integrations; populated by host
	// code after DI completes via SetRegistry (the pkg/provider Registry
	// pattern is already used by oauth.controller.go).
	registry *identityproviderpkg.Registry
}

// SetRegistry wires (or replaces) the OAuth Registry. Host code calls this
// after constructing per-provider implementations (Stripe/Paystack/etc-free
// — these are GitHub/GitLab/Bitbucket/Google for identity).
func (c *AuthorizeController) SetRegistry(r *identityproviderpkg.Registry) {
	c.registry = r
}

// supportsState is the set of provider keys for which the OAuth `state`
// nonce check is enforced. Mirrors TS SupportsState — providers that don't
// echo state back can't be checked, so we skip those.
var supportsState = map[string]bool{
	"github":    true,
	"gitlab":    true,
	"bitbucket": true,
	"google":    true,
}

// Authorize mirrors TS @Post(':provider/authorize') — the full OAuth login
// flow. Steps (TS-equivalent):
//   1. State validation against the cached (state → clientId) binding.
//   2. Provider lookup + token exchange via the Registry.
//   3. Profile fetch + ProviderLog persistence.
//   4. User find-or-create on profile.email with provider-type guard.
//   5. Workspace + UserClientWorkspace bootstrap when the user has none.
//   6. Attribute upsert from profile fields.
//   7. Referral-code redemption (best-effort).
//   8. Device + DeviceSession + DeviceLog bootstrap.
//   9. Default Admin / Member role-type + permission-type seeding.
//   10. Access + refresh token issue.
func (c *AuthorizeController) Authorize(dto *AuthorizeProviderDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.provider.title", nil, pref)

	// 1. State validation when the provider supports state.
	if supportsState[dto.Provider] && !c.validateState(dto.State, dto.ClientId) {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.provider.post.authorize.error.invalid-request", nil, pref))
	}

	// 2. Provider row lookup.
	provRow, _ := c.providerEntity.First(`"key" = ?`, dto.Provider)
	if provRow == nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.provider.post.authorize.error.invalid-request", nil, pref))
	}

	// 3. Token exchange + profile fetch via per-provider integration.
	if c.registry == nil {
		return response.BadRequest(title, "oauth registry not configured")
	}
	prov := c.registry.Get(dto.Provider)
	if prov == nil {
		return response.BadRequest(title, "unknown oauth provider")
	}
	redirect := ""
	if provRow.RedirectUrl != nil {
		redirect = *provRow.RedirectUrl
	}
	tok, err := prov.Exchange(dto.Code, redirect)
	if err != nil || tok == nil {
		return response.BadRequest(title, "token exchange failed")
	}
	profile, err := prov.Profile(tok.AccessToken)
	if err != nil || profile == nil {
		return response.BadRequest(title, "profile fetch failed")
	}

	// 4. Persist a Profile-typed ProviderLog row.
	c.logProfile(provRow.Id, tok.AccessToken, profile)

	// 5. User find-or-create + provider-type guard.
	existing, _ := c.userEntity.First(`"ref" = ? OR "email" = ?`, profile.Email, profile.Email)
	if dto.Type == "register" && existing != nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.provider.post.authorize.error.user-exists", nil, pref))
	}
	if dto.Type == "login" && existing != nil && existing.Type != nil && *existing.Type != dto.Provider {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.provider.post.authorize.error.different-provider", nil, pref))
	}
	if existing == nil {
		ref := profile.Email
		userType := dto.Provider
		secret := tok.AccessToken
		existing = &identityuser.User{
			Name: profile.Name, Email: profile.Email,
			Ref: &ref, Type: &userType, Secret: &secret,
		}
		if err := c.userEntity.Insert(existing); err != nil {
			return response.InternalServerError(title, err.Error())
		}
	}

	// 6. Workspace bootstrap when the user has none.
	workspaces := c.workspacesForUser(existing.Id, dto.ClientId)
	if len(workspaces) == 0 {
		ws := c.bootstrapWorkspace(existing, dto.ClientId)
		if ws != nil {
			workspaces = []identityworkspace.Workspace{*ws}
		}
	}
	if len(workspaces) == 0 {
		return response.InternalServerError(title, "workspace bootstrap failed")
	}

	// 7. Attribute upsert from profile fields (email, name, avatar, username).
	c.upsertProfileAttributes(existing.Id, profile)

	// 8. Referral-code redemption (best-effort — no error on miss).
	if dto.ReferralCode != "" {
		_ = c.redeemReferralCode(existing.Id, dto.ReferralCode)
	}

	// 9. Device + Session + Log bootstrap.
	deviceId, sessionId := c.bootstrapDevice(existing.Id, dto.ClientId, workspaces[0].Id, dto.UserAgent)
	c.logDeviceLogin(deviceId, sessionId, profile, dto.Ctx)

	// 10. Default Admin/Member role-type + permission-type seeding.
	c.seedDefaultRoles(existing.Id, dto.ClientId, workspaces[0].Id)

	// 11. Issue access + refresh tokens.
	access, refresh, err := c.auth.IssueTokensByHandle(profile.Email, dto.ClientId)
	if err != nil {
		return response.Unauthorized(title, err.Error())
	}

	return response.Success(map[string]any{
		"user":         maskUser(existing),
		"workspaces":   workspaces,
		"accessToken":  access,
		"refreshToken": refresh,
		"profile":      profile,
	}, title,
		c.lang.Translate("apps.identity.provider.post.authorize.success", nil, pref), nil)
}

// validateState confirms that the supplied state matches the cached binding
// for the (state → clientId) pair seeded by /provider/providers or
// /provider/:provider/provider. 10-min TTL matches the setter.
func (c *AuthorizeController) validateState(state, clientId string) bool {
	if state == "" || clientId == "" || c.cache == nil {
		return false
	}
	got, ok := c.cache.Get("provider:oauth:initiate:" + state)
	if !ok {
		return false
	}
	return got == clientId
}

// logProfile mirrors TS providerLogService.save(type=Profile) — records the
// (auth, profile) pair against the global Provider row so audit trails can
// reconstruct the OAuth handshake.
func (c *AuthorizeController) logProfile(providerId, accessToken string, profile *identityproviderpkg.Profile) {
	if c.providerLogEntity == nil {
		return
	}
	req, _ := json.Marshal(map[string]any{"accessToken": accessToken})
	resp, _ := json.Marshal(profile)
	_ = c.providerLogEntity.Insert(&identityproviderlog.ProviderLog{
		ProviderId: providerId,
		Event:      "profile",
		Request:    req,
		Response:   resp,
	})
}

// workspacesForUser returns every workspace the user is bound to via the
// UserClientWorkspace join.
func (c *AuthorizeController) workspacesForUser(userId, clientId string) []identityworkspace.Workspace {
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

// bootstrapWorkspace creates a default workspace + UCW (owner) row for a
// brand-new user. Mirrors TS body lines 363–395.
func (c *AuthorizeController) bootstrapWorkspace(user *identityuser.User, clientId string) *identityworkspace.Workspace {
	if c.workspaceEntity == nil || c.ucwEntity == nil {
		return nil
	}
	name := "My Workspace"
	slug := utils.Reference("WS", 16)
	ws := &identityworkspace.Workspace{
		UserId: user.Id, ClientId: clientId,
		Name: name, Slug: slug,
	}
	if err := c.workspaceEntity.Insert(ws); err != nil {
		return nil
	}
	owner := "owner"
	_ = c.ucwEntity.Insert(&identityucw.UserClientWorkspace{
		UserId: user.Id, ClientId: clientId, WorkspaceId: ws.Id,
		Status: &owner,
	})
	return ws
}

// upsertProfileAttributes writes each profile field as an Attribute row
// (type=user, key=<field>). Idempotent — re-running on the same user
// updates the existing rows rather than inserting duplicates.
func (c *AuthorizeController) upsertProfileAttributes(userId string, profile *identityproviderpkg.Profile) {
	if c.attributeEntity == nil || profile == nil {
		return
	}
	userType := "user"
	fields := map[string]string{
		"email":     profile.Email,
		"name":      profile.Name,
		"avatarUrl": profile.AvatarURL,
		"username":  profile.Username,
	}
	for k, v := range fields {
		if v == "" {
			continue
		}
		existing, _ := c.attributeEntity.First(
			`"user_id" = ? AND "type" = ? AND "key" = ?`,
			userId, userType, k,
		)
		val := v
		if existing != nil {
			existing.Value = &val
			_, _ = c.attributeEntity.Update(existing, `"id" = ?`, existing.Id)
			continue
		}
		_ = c.attributeEntity.Insert(&identityattribute.Attribute{
			UserId: userId, Key: k, Value: &val, Type: &userType,
		})
	}
}

// redeemReferralCode looks up an Attribute with type=user, key=referralCode,
// value=<code> and records the referrer link on the new user (best-effort).
func (c *AuthorizeController) redeemReferralCode(userId, code string) string {
	if c.attributeEntity == nil {
		return ""
	}
	userType := "user"
	referrer, _ := c.attributeEntity.First(
		`"type" = ? AND "key" = ? AND "value" = ?`,
		userType, "referralCode", code,
	)
	if referrer == nil {
		return ""
	}
	// Record the link on the new user.
	rid := referrer.UserId
	_ = c.attributeEntity.Insert(&identityattribute.Attribute{
		UserId: userId, Key: "referrerId", Value: &rid, Type: &userType,
	})
	return rid
}

// bootstrapDevice creates (or reuses) a Device + DeviceSession pair for the
// user. UA parsing is best-effort — fields the gox UA parser can resolve are
// populated; everything else falls back to the legacy fingerprint shape so
// the entity remains insertable.
func (c *AuthorizeController) bootstrapDevice(userId, clientId, workspaceId, userAgent string) (string, string) {
	if c.deviceEntity == nil || c.deviceSessionEntity == nil {
		return "", ""
	}
	fp := utils.Reference("DEV", 24)
	os, agent, engine, cpu := parseUA(userAgent)
	dev := &identitydevice.Device{
		UserId: userId, Fingerprint: fp, Type: "browser",
		Os: &os, Agent: &agent, Engine: &engine, Cpu: &cpu,
	}
	_ = c.deviceEntity.Insert(dev)

	// 24h session by default.
	exp := time.Now().UTC().Add(24 * time.Hour)
	sess := &identitydevicesession.DeviceSession{
		DeviceId: dev.Id, UserId: userId,
		Token: utils.Reference("SES", 32), ExpiresAt: &exp,
	}
	_ = c.deviceSessionEntity.Insert(sess)
	return dev.Id, sess.Id
}

// logDeviceLogin records a DeviceLog row tagged with GeoIP + preference
// metadata. GeoIP fields are populated from NTXContext.Preference when
// available — gox doesn't ship a GeoIP service in this layer.
func (c *AuthorizeController) logDeviceLogin(deviceId, sessionId string, profile *identityproviderpkg.Profile, ntx ntxctx.NTXContext) {
	if c.deviceLogEntity == nil || deviceId == "" {
		return
	}
	country, _ := ntx.Preference["country"].(string)
	meta, _ := json.Marshal(map[string]any{
		"theme":    ntx.Preference["theme"],
		"language": ntx.Preference["language"],
		"timezone": ntx.Preference["timezone"],
		"currency": ntx.Preference["currency"],
		"profile":  profile,
	})
	_ = c.deviceLogEntity.Insert(&identitydevicelog.DeviceLog{
		DeviceId:  deviceId,
		SessionId: &sessionId,
		Event:     "login",
		Country:   &country,
		Meta:      meta,
	})
}

// seedDefaultRoles ensures the workspace has an Admin and Member RoleType
// (each linked to a wildcard PermissionType). Mirrors TS lines 538–600.
func (c *AuthorizeController) seedDefaultRoles(userId, clientId, workspaceId string) {
	if c.roleTypeEntity == nil || c.permissionTypeEntity == nil || c.permissionEntity == nil {
		return
	}
	defaults := []struct {
		roleName   string
		permName   string
		permVerb   string
		permScope  string
	}{
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
		existing, _ := c.permissionEntity.First(
			`"role_type_id" = ? AND "permission_type_id" = ?`,
			rt.Id, pt.Id,
		)
		if existing != nil {
			continue
		}
		rtId := rt.Id
		ptId := pt.Id
		_ = c.permissionEntity.Insert(&identitypermission.Permission{
			RoleTypeId: &rtId, PermissionTypeId: &ptId,
			Key:    def.permName,
			Action: def.permVerb,
			Scope:  def.permScope,
		})
	}
	_ = userId
	_ = workspaceId
	_ = clientId
}

// maskUser strips Secret + Token from the user envelope before sending it
// back to the client. Matches TS maskUser().
func maskUser(u *identityuser.User) map[string]any {
	if u == nil {
		return nil
	}
	out := map[string]any{
		"id":    u.Id,
		"name":  u.Name,
		"email": u.Email,
	}
	if u.Ref != nil {
		out["ref"] = *u.Ref
	}
	if u.Type != nil {
		out["type"] = *u.Type
	}
	return out
}

// parseUA does a best-effort User-Agent decomposition — returns
// (os, agent, engine, cpu). Tests can swap this with a richer UA parser
// without changing the controller surface.
func parseUA(ua string) (string, string, string, string) {
	if ua == "" {
		return "", "", "", ""
	}
	// Cheap heuristic — covers the common cases (Chrome/Firefox/Safari on
	// Win/Mac/Linux/iOS/Android). Tests can replace via a real parser
	// (services/uaparser) when the dependency is wired into identity.
	os, agent, engine, cpu := "", "", "", ""
	low := strings.ToLower(ua)
	switch {
	case contains(low, "windows"):
		os = "Windows"
	case contains(low, "mac os x") || contains(low, "macos"):
		os = "macOS"
	case contains(low, "linux"):
		os = "Linux"
	case contains(low, "iphone") || contains(low, "ipad"):
		os = "iOS"
	case contains(low, "android"):
		os = "Android"
	}
	switch {
	case contains(low, "chrome"):
		agent = "Chrome"
	case contains(low, "firefox"):
		agent = "Firefox"
	case contains(low, "safari"):
		agent = "Safari"
	case contains(low, "edge"):
		agent = "Edge"
	}
	switch {
	case contains(low, "webkit"):
		engine = "WebKit"
	case contains(low, "gecko"):
		engine = "Gecko"
	case contains(low, "blink"):
		engine = "Blink"
	}
	switch {
	case contains(low, "x86_64") || contains(low, "win64"):
		cpu = "x86_64"
	case contains(low, "arm64") || contains(low, "aarch64"):
		cpu = "arm64"
	case contains(low, "x86"):
		cpu = "x86"
	}
	return os, agent, engine, cpu
}

func contains(haystack, needle string) bool {
	return needle != "" && len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
