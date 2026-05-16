package app

import (
	"encoding/json"
	"time"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationprovider "github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
	queueapp "github.com/thescaffold/gox-apps/libs/queue/app"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/services"
)

// AppController mirrors ntx-apps/libs/notification/src/app.controller.ts.
// Endpoints:
//
//	GET    /                getHello
//	GET    /scope           findByScope
//	GET    /priority        findByPriority
//	GET    /subscription    getSubscription
//	PATCH  /subscription    updateSubscription
type AppController struct {
	appService *AppService                          `inject:""`
	provider   *notificationprovider.ProviderService `inject:""`
	queue      *queueapp.AppService                  `inject:""`
	cache      services.CacheBackend                 `inject:""`
	lang       *i18n.Service                         `inject:""`
}

// OnRegister fires after DI completes; wires the queueapp-backed pusher
// into the ProviderService so AppService.OnMessageNew dispatches durably
// rather than running Send synchronously on the request goroutine.
func (c *AppController) OnRegister() {
	if c.provider == nil || c.queue == nil {
		return
	}
	c.provider.SetQueuePusher(&queuePump{queue: c.queue})
}

// hideKeyTTL is the duration the AppController marks a per-user message as
// "already-shown" in cache. Long enough that the same user paging through
// won't see the same message twice in a session, short enough that a stale
// cache entry doesn't permanently hide content. TS doesn't set an explicit
// TTL (it inherits the CacheService default); gox picks 24h here.
const hideKeyTTL = 24 * time.Hour

// flattenWithHide produces the TS-shape `{...message, hide:bool}` envelope.
// We JSON-roundtrip the entity into a map so callers can extend it with the
// hide flag — mirrors `{...message, hide: !!alreadyShown}` exactly.
func flattenWithHide(item any, hide bool) map[string]any {
	out := map[string]any{}
	if b, err := json.Marshal(item); err == nil {
		_ = json.Unmarshal(b, &out)
	}
	out["hide"] = hide
	return out
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — plain string,
// no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// FindByScope mirrors TS @Get('scope') findByScope. Missing scope/channel
// returns success([]) with the apps.notification.app.get.message.scope.error
// message (i.e. behaves as a "soft" validation per TS). Each returned item
// is enriched with a `hide` flag based on a per-user message-seen cache key.
func (c *AppController) FindByScope(dto *FindByScopeDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.notification.app.title", nil, pref)
	if dto.Scope == "" || dto.Channel == "" {
		return response.Success([]any{},
			title,
			c.lang.Translate("apps.notification.app.get.message.scope.error", nil, pref),
			nil)
	}
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	out, err := c.appService.FindByScope(ScopeQuery{
		Channel: dto.Channel, Scope: dto.Scope,
		UserID: userID, ClientID: clientID, WorkspaceID: workspaceID,
		Page: dto.Page, PerPage: dto.PerPage,
	})
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	items := c.applyHideTracking(out.Items, userID, workspaceID)
	return response.Paginated(items, out.Page, out.PerPage, out.Total,
		title,
		c.lang.Translate("apps.notification.app.get.message.success", nil, pref),
	)
}

// FindByPriority mirrors TS @Get('priority') findByPriority. Same soft-validation
// behaviour as FindByScope when priority/channel are missing.
func (c *AppController) FindByPriority(dto *FindByPriorityDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.notification.app.title", nil, pref)
	if dto.Priority == "" || dto.Channel == "" {
		return response.Success([]any{},
			title,
			c.lang.Translate("apps.notification.app.get.message.scope.error", nil, pref),
			nil)
	}
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	out, err := c.appService.FindByPriority(PriorityQuery{
		Channel: dto.Channel, Priority: dto.Priority,
		UserID: userID, ClientID: clientID, WorkspaceID: workspaceID,
		Page: dto.Page, PerPage: dto.PerPage,
	})
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	items := c.applyHideTracking(out.Items, userID, workspaceID)
	return response.Paginated(items, out.Page, out.PerPage, out.Total,
		title,
		c.lang.Translate("apps.notification.app.get.message.success", nil, pref),
	)
}

// applyHideTracking mirrors the TS per-user "already shown" cache so paginating
// the same query within hideKeyTTL doesn't re-surface seen messages. The cache
// key matches TS: apps:notification:messages:<userId>:<workspaceId>:<messageId>.
// First-seen messages return hide:false; subsequent reads return hide:true.
// When the cache is unwired (nil), every message returns hide:false (parity
// with TS when CacheService is absent).
func (c *AppController) applyHideTracking(items []message.Message, userID, workspaceID string) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for i := range items {
		m := &items[i]
		hide := false
		if c.cache != nil {
			// Setnx is atomic: returns true when this caller was the first to
			// write the key. Subsequent (or concurrent) callers see ok=false
			// and hide=true — replica-safe coordination without a separate
			// Get/Set race window.
			key := "apps:notification:messages:" + userID + ":" + workspaceID + ":" + m.Id
			firstSeen := c.cache.Setnx(key, "1", hideKeyTTL)
			hide = !firstSeen
		}
		out = append(out, flattenWithHide(m, hide))
	}
	return out
}

// GetSubscription mirrors TS @Get('subscription') getSubscription. Missing ref
// returns success(null) with the ref-missing translation.
func (c *AppController) GetSubscription(dto *GetSubscriptionDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.notification.app.title", nil, pref)
	if dto.Ref == "" {
		return response.Success(nil,
			title,
			c.lang.Translate("apps.notification.subscription.get.error.ref-missing", nil, pref),
			nil)
	}
	r, _ := c.appService.GetSubscription(dto.Ref)
	return response.Success(r,
		title,
		c.lang.Translate("apps.notification.subscription.get.success", nil, pref),
		nil)
}

// UpdateSubscription mirrors TS @Patch('subscription') updateSubscription.
// Missing ref → error envelope with ref-missing translation.
func (c *AppController) UpdateSubscription(dto *UpdateSubscriptionDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.notification.app.title", nil, pref)
	if dto.Ref == "" {
		return response.BadRequest(title,
			c.lang.Translate("apps.notification.subscription.update.error.ref-missing", nil, pref),
		)
	}
	r, err := c.appService.UpdateSubscription(dto.Ref, dto.Rules)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(r,
		title,
		c.lang.Translate("apps.notification.subscription.update.success", nil, pref),
		nil)
}

// identityFromContext extracts user/client/workspace ids from NTXContext.
func identityFromContext(ntx ntxctx.NTXContext) (string, string, string) {
	var userID, clientID, workspaceID string
	if ntx.User != nil {
		userID, _ = ntx.User["id"].(string)
	}
	if userID == "" {
		userID = ntx.UserID
	}
	if ntx.Client != nil {
		clientID, _ = ntx.Client["id"].(string)
	}
	if clientID == "" {
		clientID = ntx.ClientID
	}
	if ntx.Workspace != nil {
		workspaceID, _ = ntx.Workspace["id"].(string)
	}
	if workspaceID == "" {
		workspaceID = ntx.WorkspaceID
	}
	return userID, clientID, workspaceID
}
