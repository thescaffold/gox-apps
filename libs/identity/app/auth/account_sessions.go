package auth

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/awesome-goose/goose/types"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/pushtoken"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Account › Sessions (V0-15, PLT-02): a person's own device sessions on every device, each with
// its browser, place and last use, and signing one out, or every one but this. Only the caller's
// own sessions are ever listed or ended; another person's session is "not found". A session's
// tokens name it (sid), so signing it out ends them at once (pkg/revocation), and the device can
// no longer re-mint (GET /sessions).
//
//	GET    /account/sessions                 the caller's sessions, this one marked current
//	DELETE /account/sessions/:id             sign one of them out
//	POST   /account/sessions/revoke-others   sign out everywhere else

// AccountSessionsDto carries GET /account/sessions and POST /account/sessions/revoke-others.
type AccountSessionsDto struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Cookie string            `header:"Cookie"`
}

// AccountSessionDto carries DELETE /account/sessions/:id.
type AccountSessionDto struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Cookie string            `header:"Cookie"`
	Id     string            `param:"id"`
}

// PlaceResolver names the place an address is in ("Berlin, Germany"), for the sessions list. The
// host registers one (a GeoIP database); without it a session shows its address.
type PlaceResolver func(ip string) string

var placeResolver struct {
	mu sync.RWMutex
	fn PlaceResolver
}

// SetPlaceResolver installs the host's PlaceResolver; nil removes it.
func SetPlaceResolver(fn PlaceResolver) {
	placeResolver.mu.Lock()
	placeResolver.fn = fn
	placeResolver.mu.Unlock()
}

func placeOf(ip string) string {
	placeResolver.mu.RLock()
	fn := placeResolver.fn
	placeResolver.mu.RUnlock()
	if fn == nil || ip == "" {
		return ""
	}
	return fn(ip)
}

// seenEvery bounds how often a session's last use is written.
const seenEvery = 5 * time.Minute

// touchSession records that a session was just used, from ip ("" keeps the address it had).
func (c *AuthController) touchSession(row *identitydevicesession.DeviceSession, ip string) {
	if c.db == nil || c.db.DB == nil || row == nil || row.Id == "" {
		return
	}
	now := time.Now().UTC()
	if row.LastSeenAt != nil && now.Sub(*row.LastSeenAt) < seenEvery && (ip == "" || ip == row.IP) {
		return
	}
	set := map[string]any{"last_seen_at": now}
	if ip != "" && ip != row.IP {
		set["ip"], set["place"] = ip, placeOf(ip)
	}
	_ = c.db.DB.Table(`"IdentityDeviceSessions"`).Where(`"id" = ?`, row.Id).Updates(set).Error
}

var (
	reEdge    = regexp.MustCompile(`Edg(e|A|iOS)?/`)
	reFirefox = regexp.MustCompile(`(Firefox|FxiOS)/`)
	reChrome  = regexp.MustCompile(`(Chrome|CriOS)/`)
	reSafari  = regexp.MustCompile(`Version/[\d.]+.*Safari/`)
)

// describeAgent reads a browser and an operating system out of a User-Agent ("Chrome", "macOS").
func describeAgent(ua string) (browser, os string) {
	switch {
	case ua == "":
	case reEdge.MatchString(ua):
		browser = "Edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera"):
		browser = "Opera"
	case reFirefox.MatchString(ua):
		browser = "Firefox"
	case reChrome.MatchString(ua):
		browser = "Chrome"
	case reSafari.MatchString(ua):
		browser = "Safari"
	}
	switch {
	case ua == "":
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") || strings.Contains(ua, "iPod"):
		os = "iOS"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "Mac OS X") || strings.Contains(ua, "Macintosh"):
		os = "macOS"
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "CrOS"):
		os = "ChromeOS"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	}
	return browser, os
}

// currentSession is the caller's own session id: the one its token names, else the one this
// browser's device cookie holds for them.
func (c *AuthController) currentSession(ntx ntxctx.NTXContext, cookie string) string {
	if sid := sessionOf(ntx); sid != "" {
		return sid
	}
	device := cookieValue(cookie, DeviceCookie)
	if device == "" || c.deviceSessionEntity == nil {
		return ""
	}
	if row, _ := c.deviceSessionEntity.First(`"device_id" = ? AND "user_id" = ?`, device, ntx.UserID); row != nil {
		return row.Id
	}
	return ""
}

// AccountSessions is GET /account/sessions.
func (c *AuthController) AccountSessions(dto *AccountSessionsDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if dto.Ctx.UserID == "" || c.deviceSessionEntity == nil {
		return response.Unauthorized(title, "sign in to see your sessions")
	}
	rows, err := c.deviceSessionEntity.Find(0, 0,
		`"user_id" = ? AND ("expires_at" IS NULL OR "expires_at" > timezone('UTC', now()))`, dto.Ctx.UserID)
	if err != nil {
		return response.InternalServerError(title, "could not list your sessions")
	}
	current := c.currentSession(dto.Ctx, dto.Cookie)
	agents := map[string]string{}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		agent, ok := agents[row.DeviceId]
		if !ok && c.deviceEntity != nil {
			if d, _ := c.deviceEntity.First(`"id" = ?`, row.DeviceId); d != nil && d.Agent != nil {
				agent = *d.Agent
			}
			agents[row.DeviceId] = agent
		}
		browser, os := describeAgent(agent)
		seen := row.LastSeenAt
		if seen == nil {
			seen = row.UpdatedAt
		}
		out = append(out, map[string]any{
			"id": row.Id, "current": row.Id == current, "browser": browser, "os": os,
			"ip": row.IP, "place": row.Place, "lastSeenAt": seen, "createdAt": row.CreatedAt,
			"expiresAt": row.ExpiresAt, "sso": row.SsoConnectionId != "",
		})
	}
	stamp := func(m map[string]any) time.Time {
		if t, ok := m["lastSeenAt"].(*time.Time); ok && t != nil {
			return *t
		}
		return time.Time{}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i]["current"] != out[j]["current"] {
			return out[i]["current"] == true
		}
		return stamp(out[i]).After(stamp(out[j]))
	})
	return response.Success(out, title, "your sessions", nil)
}

// endSessions signs the given sessions of userID out: their rows go (so their tokens are refused
// and the devices cannot re-mint) and their devices' push tokens with them.
func (c *AuthController) endSessions(userID string, rows []identitydevicesession.DeviceSession) error {
	if len(rows) == 0 {
		return nil
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	if _, err := c.deviceSessionEntity.Delete(`"user_id" = ? AND "id" IN ?`, userID, ids); err != nil {
		return err
	}
	if c.db != nil && c.db.DB != nil {
		push := pushtoken.New(c.db.DB)
		for _, r := range rows {
			_, _ = push.DeregisterDevice(context.Background(), userID, r.DeviceId)
		}
	}
	return nil
}

// AccountSessionRevoke is DELETE /account/sessions/:id: sign one of one's own sessions out.
func (c *AuthController) AccountSessionRevoke(dto *AccountSessionDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if dto.Ctx.UserID == "" || c.deviceSessionEntity == nil {
		return response.Unauthorized(title, "sign in to manage your sessions")
	}
	row, _ := c.deviceSessionEntity.First(`"id" = ? AND "user_id" = ?`, dto.Id, dto.Ctx.UserID)
	if row == nil {
		return response.NotFound(title, "no such session")
	}
	if err := c.endSessions(dto.Ctx.UserID, []identitydevicesession.DeviceSession{*row}); err != nil {
		return response.InternalServerError(title, "could not sign that session out")
	}
	return response.Success(map[string]any{"id": row.Id, "ended": 1, "current": row.Id == c.currentSession(dto.Ctx, dto.Cookie)}, title, "signed out", nil)
}

// AccountSessionsRevokeOthers is POST /account/sessions/revoke-others: sign out everywhere but
// here. "Here" is the session the token names, or this browser's; with neither it is refused,
// since it would sign out every session.
func (c *AuthController) AccountSessionsRevokeOthers(dto *AccountSessionsDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if dto.Ctx.UserID == "" || c.deviceSessionEntity == nil {
		return response.Unauthorized(title, "sign in to manage your sessions")
	}
	keep := c.currentSession(dto.Ctx, dto.Cookie)
	if keep == "" {
		return response.BadRequest(title, "this request belongs to no session; sign out everywhere instead")
	}
	rows, err := c.deviceSessionEntity.Find(0, 0, `"user_id" = ? AND "id" <> ?`, dto.Ctx.UserID, keep)
	if err != nil {
		return response.InternalServerError(title, "could not sign the other sessions out")
	}
	if err := c.endSessions(dto.Ctx.UserID, rows); err != nil {
		return response.InternalServerError(title, "could not sign the other sessions out")
	}
	return response.Success(map[string]any{"ended": len(rows), "kept": keep}, title, "signed out everywhere else", nil)
}
