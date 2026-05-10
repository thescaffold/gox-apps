package app

import (
	"github.com/awesome-goose/goose/types"
	ntxctx "github.com/thescaffold/gox-packages-core/context"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

// Health is the GET / handler.
func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "notification", "ok", nil)
}

// FindByScope handles GET /scope — paginated messages filtered by (channel, scope).
func (c *AppController) FindByScope(ctx types.Context, dto *FindByScopeDto) types.Output {
	ntx := ntxctx.Get(ctx)
	userID, clientID, workspaceID := identityFromContext(ntx)
	if clientID == "" {
		return response.BadRequest("notification", "client context required")
	}
	out, err := c.appService.FindByScope(ScopeQuery{
		Channel: dto.Channel, Scope: dto.Scope,
		UserID: userID, ClientID: clientID, WorkspaceID: workspaceID,
		Page: dto.Page, PerPage: dto.PerPage,
	})
	if err != nil {
		return response.InternalServerError("notification", err.Error())
	}
	return response.Paginated(out.Items, out.Page, out.PerPage, out.Total, "notification", "ok")
}

// FindByPriority handles GET /priority — paginated messages filtered by (channel, priority).
func (c *AppController) FindByPriority(ctx types.Context, dto *FindByPriorityDto) types.Output {
	ntx := ntxctx.Get(ctx)
	userID, clientID, workspaceID := identityFromContext(ntx)
	if clientID == "" {
		return response.BadRequest("notification", "client context required")
	}
	out, err := c.appService.FindByPriority(PriorityQuery{
		Channel: dto.Channel, Priority: dto.Priority,
		UserID: userID, ClientID: clientID, WorkspaceID: workspaceID,
		Page: dto.Page, PerPage: dto.PerPage,
	})
	if err != nil {
		return response.InternalServerError("notification", err.Error())
	}
	return response.Paginated(out.Items, out.Page, out.PerPage, out.Total, "notification", "ok")
}

// GetSubscription handles GET /subscription?ref=…
func (c *AppController) GetSubscription(dto *GetSubscriptionDto) types.Output {
	r, err := c.appService.GetSubscription(dto.Ref)
	if err != nil {
		return response.BadRequest("notification", err.Error())
	}
	return response.Success(r, "notification", "ok", nil)
}

// UpdateSubscription handles PATCH /subscription.
func (c *AppController) UpdateSubscription(dto *UpdateSubscriptionDto) types.Output {
	r, err := c.appService.UpdateSubscription(dto.Ref, dto.Rules)
	if err != nil {
		return response.BadRequest("notification", err.Error())
	}
	return response.Success(r, "notification", "ok", nil)
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
