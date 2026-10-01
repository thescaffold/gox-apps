package auth

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/awesome-goose/goose/types"
	identityclient "github.com/thescaffold/gox-apps/libs/identity/app/client"
	identityclientlog "github.com/thescaffold/gox-apps/libs/identity/app/clientlog"
	identitydevice "github.com/thescaffold/gox-apps/libs/identity/app/device"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// The browser shell is built around a device: a cookie names it, the device
// holds one session per signed-in account, and the shell asks for those
// sessions (GET sessions) to show the accounts, to switch between them and to
// get a fresh token when the short-lived one runs out. This file is that
// surface (PLAN M1-40); the sign-in handlers in auth.controller.go create the
// device and set the cookie.

// DeviceCookie is the cookie that names a browser's device.
const DeviceCookie = "x_ntx_device_id"

// deviceLife is how long a device session, and its cookie, last.
const deviceLife = 30 * 24 * time.Hour

// cookieValue returns the value of one cookie in a Cookie header, or "".
func cookieValue(header, name string) string {
	for _, part := range strings.Split(header, ";") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && k == name {
			return v
		}
	}
	return ""
}

// deviceCookie builds the Set-Cookie value for a device. It is HttpOnly (the
// page's scripts never read it), SameSite=Lax (a cross-site request does not
// carry it) and Secure when the site is served over https.
func deviceCookie(deviceID string, life time.Duration) string {
	secure := ""
	if strings.HasPrefix(strings.ToLower(os.Getenv("BASE_URL")), "https://") {
		secure = "; Secure"
	}
	return DeviceCookie + "=" + deviceID + "; Path=/; Max-Age=" + strconv.Itoa(int(life.Seconds())) + "; HttpOnly; SameSite=Lax" + secure
}

// withDeviceCookie puts the device cookie on a response.
func withDeviceCookie(out types.Output, deviceID string) types.Output {
	if deviceID != "" {
		if h := out.Headers(); h != nil {
			h["Set-Cookie"] = deviceCookie(deviceID, deviceLife)
		}
	}
	return out
}

// bootstrapDevice finds or creates the browser's device and makes sure it holds
// a live session for this user. A device id the caller names is trusted only
// if that device exists: an id that matches nothing is ignored and a new device
// is made, so a forged cookie cannot be used to attach sessions to someone else's
// device, or to read a device that was never issued.
func (c *AuthController) bootstrapDevice(userId, clientId, userAgent, wantDevice string) (deviceID, sessionID string) {
	if c.deviceEntity == nil || c.deviceSessionEntity == nil {
		return "", ""
	}
	var dev *identitydevice.Device
	if wantDevice != "" {
		if d, err := c.deviceEntity.First(`"id" = ?`, wantDevice); err == nil && d != nil {
			dev = d
		}
	}
	if dev == nil {
		dev = &identitydevice.Device{UserId: userId, Fingerprint: utils.Reference("DEV", 24), Type: "browser"}
		if userAgent != "" {
			dev.Agent = &userAgent
		}
		if err := c.deviceEntity.Insert(dev); err != nil {
			return "", ""
		}
	}
	exp := time.Now().UTC().Add(deviceLife)
	client := clientId
	var cp *string
	if client != "" {
		cp = &client
	}
	if existing, _ := c.deviceSessionEntity.First(`"device_id" = ? AND "user_id" = ?`, dev.Id, userId); existing != nil {
		// signing in again renews the session (and its updated time, which orders the list)
		_, _ = c.deviceSessionEntity.Update(&identitydevicesession.DeviceSession{ExpiresAt: &exp, ClientId: cp}, `"id" = ?`, existing.Id)
		return dev.Id, existing.Id
	}
	sess := &identitydevicesession.DeviceSession{DeviceId: dev.Id, UserId: userId, Token: utils.Reference("SES", 32), ExpiresAt: &exp, ClientId: cp}
	if err := c.deviceSessionEntity.Insert(sess); err != nil {
		return dev.Id, ""
	}
	return dev.Id, sess.Id
}

// usable reports whether an account may be signed in: not closed, not suspended.
func (c *AuthController) usable(userID string) bool {
	if c.attributeEntity == nil {
		return true
	}
	for _, key := range []string{"closed_at", "suspended_at"} {
		if a, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, userID, key); a != nil {
			return false
		}
	}
	return true
}

// Sessions is GET /sessions: the accounts this browser's device is signed in to,
// most recent first, each with its workspaces and a token for each one. It is
// answered from the device cookie alone. A browser with no cookie, or one that
// names no device, has no sessions; a session that has expired, or whose
// account was closed or suspended, is left out.
//
// Only access tokens are minted, and nothing is written: the device session is
// the long-lived credential, the token is what it hands out.
func (c *AuthController) Sessions(dto *SessionsDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	ok := func(v any) types.Output {
		return response.Success(v, title, c.lang.Translate("apps.identity.app.get.sessions.success", nil, pref), nil)
	}

	deviceID := cookieValue(dto.Cookie, DeviceCookie)
	if deviceID == "" || c.deviceSessionEntity == nil {
		return ok([]any{})
	}
	rows, err := c.deviceSessionEntity.Find(0, 0,
		`"device_id" = ? AND ("expires_at" IS NULL OR "expires_at" > timezone('UTC', now()))`, deviceID)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}

	type session struct {
		updated time.Time
		out     map[string]any
	}
	stamp := func(t *time.Time) time.Time {
		if t == nil {
			return time.Time{}
		}
		return *t
	}
	var sessions []session
	for _, row := range rows {
		u, _ := c.userEntity.First(`"id" = ?`, row.UserId)
		if u == nil || !c.usable(u.Id) {
			continue
		}
		clientID := dto.Ctx.ClientID
		if row.ClientId != nil && *row.ClientId != "" {
			clientID = *row.ClientId
		}
		sessions = append(sessions, session{
			updated: stamp(row.UpdatedAt),
			out: map[string]any{
				"id": row.Id, "deviceId": row.DeviceId, "userId": row.UserId, "clientId": clientID,
				"updatedAt": row.UpdatedAt, "user": c.sessionUser(u, clientID),
			},
		})
	}
	sort.SliceStable(sessions, func(i, j int) bool { return sessions[i].updated.After(sessions[j].updated) })
	out := make([]any, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, s.out)
	}
	return ok(out)
}

// sessionUser is one account as the shell shows it: who they are, their
// workspaces each with a token for that workspace, their attributes, and their
// roles and permissions in the first workspace.
func (c *AuthController) sessionUser(u *identityuser.User, clientID string) map[string]any {
	user := maskUser(u)
	var links []any
	firstWorkspace := ""
	for _, w := range c.workspacesForUser(u.Id, clientID) {
		if firstWorkspace == "" {
			firstWorkspace = w.Id
		}
		kind := ""
		if l, _ := c.ucwEntity.First(`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ?`, u.Id, clientID, w.Id); l != nil && l.Status != nil {
			kind = *l.Status
		}
		ws := map[string]any{}
		raw, _ := json.Marshal(w)
		_ = json.Unmarshal(raw, &ws)
		if t, err := c.auth.IssueAccessToken(u, clientID, w.Id); err == nil {
			ws["token"] = t
		}
		links = append(links, map[string]any{
			"userId": u.Id, "clientId": clientID, "workspaceId": w.Id, "type": kind, "workspace": ws,
		})
	}
	if links == nil {
		links = []any{}
	}
	user["workspaces"] = links
	user["attributes"] = c.attributesForUser(u.Id)
	roles, permissions := []string{}, []string{}
	if firstWorkspace != "" {
		if rs, err := c.auth.ResolveRoles(u.Id, clientID, firstWorkspace); err == nil {
			for _, r := range rs {
				roles = append(roles, r.Name)
			}
		}
		if ps, err := c.auth.ResolvePermissions(u.Id, clientID, firstWorkspace); err == nil {
			for _, p := range ps {
				permissions = append(permissions, p.Key)
			}
		}
	}
	user["roles"], user["permissions"] = roles, permissions
	if t, err := c.auth.IssueAccessToken(u, clientID, firstWorkspace); err == nil {
		user["token"] = t
	}
	return user
}

// Logout is POST /logout: sign accounts out of a device. `device` says whose
// device ("yes" is this browser's, from its cookie; any other value longer than
// three characters is a device id), `userId` narrows to one account:
//
//	device + user  → that account on that device
//	device only    → every account on that device (and the cookie is cleared)
//	user only      → that account on every device
//	neither        → the signed-in user, everywhere
func (c *AuthController) Logout(dto *LogoutDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)

	device := ""
	switch {
	case dto.Device == "yes":
		device = cookieValue(dto.Cookie, DeviceCookie)
	case len(dto.Device) > 3:
		device = dto.Device
	}
	var where string
	var args []any
	switch {
	case device == "" && dto.UserId != "":
		where, args = `"user_id" = ?`, []any{dto.UserId}
	case device != "" && dto.UserId == "":
		where, args = `"device_id" = ?`, []any{device}
	case device != "" && dto.UserId != "":
		where, args = `"device_id" = ? AND "user_id" = ?`, []any{device, dto.UserId}
	case dto.Ctx.UserID != "":
		where, args = `"user_id" = ?`, []any{dto.Ctx.UserID}
	default:
		return response.Unauthorized(title, c.lang.Translate("apps.identity.app.post.logout.error.unauthorized", nil, pref))
	}
	if c.deviceSessionEntity != nil {
		if _, err := c.deviceSessionEntity.Delete(where, args...); err != nil {
			return response.InternalServerError(title, err.Error())
		}
	}
	out := response.Success(map[string]any{"deviceId": device}, title, c.lang.Translate("apps.identity.app.post.logout.success", nil, pref), nil)
	if device != "" && dto.UserId == "" {
		out.Headers()["Set-Cookie"] = deviceCookie("", 0)
	}
	return out
}

// commonClients are the apps every account has, listed apart from the rest.
var commonClients = map[string]bool{"id_center": true, "payments": true}

// UserClients is GET /user-clients: the apps the signed-in user can open, split
// into the ones they have used (favourites), the other ones (discover) and the
// ones everybody has (common).
func (c *AuthController) UserClients(dto *UserClientsDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	if dto.Ctx.UserID == "" {
		return response.Unauthorized(title, c.lang.Translate("apps.identity.app.get.user-clients.error.unauthorized", nil, pref))
	}
	clients, err := c.clientEntity.Find(0, 0, `1 = 1`)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	used := map[string]bool{}
	if logs, err := c.clientLogEntity.Find(0, 100, `"meta"->>'userId' = ?`, dto.Ctx.UserID); err == nil {
		for _, l := range logs {
			used[l.ClientId] = true
		}
	}
	entry := func(cl identityclient.Client) map[string]any {
		out := map[string]any{"title": cl.Name, "url": "", "desc": ""}
		if cl.Desc != nil {
			out["desc"] = *cl.Desc
		}
		return out
	}
	favorites, discover, common := []any{}, []any{}, []any{}
	for _, cl := range clients {
		switch {
		case commonClients[cl.Key]:
			common = append(common, entry(cl))
		case used[cl.Id]:
			favorites = append(favorites, entry(cl))
		default:
			discover = append(discover, entry(cl))
		}
	}
	return response.Success(map[string]any{"favorites": favorites, "discover": discover, "common": common},
		title, c.lang.Translate("apps.identity.app.get.user-clients.success", nil, pref), nil)
}

// RegisterClientLog is POST /client-log/register: the signed-in user opened an
// app. The user is always the one the token names (a userId in the body is
// ignored), and opening the same app in the same workspace again is not logged
// twice.
func (c *AuthController) RegisterClientLog(dto *ClientLogRegisterDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	if dto.Ctx.UserID == "" {
		return response.Unauthorized(title, c.lang.Translate("apps.identity.app.post.client-log.error.unauthorized", nil, pref))
	}
	meta, _ := json.Marshal(map[string]any{"userId": dto.Ctx.UserID, "workspaceId": dto.WorkspaceId, "type": dto.Type})
	existing, _ := c.clientLogEntity.First(`"client_id" = ? AND "meta"->>'userId' = ? AND COALESCE("meta"->>'workspaceId', '') = ?`,
		dto.ClientId, dto.Ctx.UserID, dto.WorkspaceId)
	if existing == nil {
		status := dto.Status
		row := &identityclientlog.ClientLog{ClientId: dto.ClientId, Event: "open", Meta: meta, Status: &status}
		if err := c.clientLogEntity.Insert(row); err != nil {
			return response.InternalServerError(title, err.Error())
		}
	}
	return response.Success(map[string]any{}, title, c.lang.Translate("apps.identity.app.post.client-log.success", nil, pref), nil)
}

