package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/awesome-goose/goose/types"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// The signed-in person's own workspaces. Both routes take the user and client
// from the verified token, never from the request, so one person cannot list
// or create workspaces for another (the generic /workspace CRUD resource
// trusts caller-supplied ids).

// maxWorkspacesPerUser caps self-service workspace creation.
const maxWorkspacesPerUser = 10

// MyWorkspacesDto carries GET /my/workspaces.
type MyWorkspacesDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// CreateMyWorkspaceDto carries POST /my/workspaces. Handle and region are
// optional: the handle defaults to one made from the name, the region to the
// host's default (pkg/org).
type CreateMyWorkspaceDto struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Name   string            `json:"name" binding:"required"`
	Handle string            `json:"handle"`
	Region string            `json:"region"`
}

// UpdateMyWorkspaceDto carries PATCH /my/workspaces/:id (owners and admins).
type UpdateMyWorkspaceDto struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Id     string            `param:"id"`
	Name   *string           `json:"name"`
	Handle *string           `json:"handle"`
}

// workspaceView is a workspace as the person's own listings show it.
func workspaceView(w *identityworkspace.Workspace, role string) map[string]any {
	v := map[string]any{"id": w.Id, "name": w.Name, "role": role}
	if w.Handle != nil {
		v["handle"] = *w.Handle
	}
	if w.Region != nil {
		v["region"] = *w.Region
	}
	return v
}

func orgError(title string, err error) types.Output {
	switch {
	case errors.Is(err, org.ErrHandleTaken):
		return response.ErrorCode(title, err.Error(), 409, "handle_taken")
	case errors.Is(err, org.ErrHandleInvalid), errors.Is(err, org.ErrRegion), errors.Is(err, org.ErrInvalid):
		return response.ErrorCode(title, strings.TrimPrefix(err.Error(), "invalid: "), 400, "invalid")
	case errors.Is(err, org.ErrNotAllowed):
		return response.Forbidden(title, "only the organization's owners and admins can change it")
	}
	return response.Internal(title, err)
}

// assignHandle gives a new workspace its handle (wanted, when free and valid,
// else one made from base) and region.
func (c *AuthController) assignHandle(ws *identityworkspace.Workspace, wanted, base, region string) error {
	if c.db == nil {
		return nil
	}
	ctx := context.Background()
	h := strings.ToLower(strings.TrimSpace(wanted))
	if h != "" {
		if err := org.ValidHandle(h); err != nil {
			return err
		}
		var n int64
		if err := c.db.DB.WithContext(ctx).Table(`"IdentityWorkspaces"`).Where(`lower("handle") = ?`, h).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return org.ErrHandleTaken
		}
	} else {
		var err error
		if h, err = org.FreeHandle(ctx, c.db.DB, org.HandleFor(base)); err != nil {
			return err
		}
	}
	r, err := org.Region(region)
	if err != nil {
		return err
	}
	ws.Handle, ws.Region = &h, &r
	return nil
}

// MyWorkspaces lists the workspaces the caller belongs to, with their role.
func (c *AuthController) MyWorkspaces(dto *MyWorkspacesDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	links, _ := c.ucwEntity.Find(0, 0, `"user_id" = ? AND "client_id" = ?`, dto.Ctx.UserID, dto.Ctx.ClientID)
	out := make([]map[string]any, 0, len(links))
	for _, l := range links {
		w, err := c.workspaceEntity.First(`"id" = ?`, l.WorkspaceId)
		if err != nil || w == nil {
			continue
		}
		role := ""
		if l.Status != nil {
			role = *l.Status
		}
		out = append(out, workspaceView(w, role))
	}
	return response.Success(out, title, "ok", nil)
}

// CreateMyWorkspace makes a new workspace owned by the caller.
func (c *AuthController) CreateMyWorkspace(dto *CreateMyWorkspaceDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	name := strings.TrimSpace(dto.Name)
	if name == "" || len([]rune(name)) > 80 {
		return response.BadRequest(title, "a workspace name is 1-80 characters")
	}
	links, _ := c.ucwEntity.Find(0, 0, `"user_id" = ? AND "client_id" = ?`, dto.Ctx.UserID, dto.Ctx.ClientID)
	if len(links) >= maxWorkspacesPerUser {
		return response.BadRequest(title, "you have reached the workspace limit")
	}
	ws := &identityworkspace.Workspace{
		UserId: dto.Ctx.UserID, ClientId: dto.Ctx.ClientID, Name: name, Slug: utils.Reference("WS", 16),
	}
	if err := c.assignHandle(ws, dto.Handle, name, dto.Region); err != nil {
		return orgError(title, err)
	}
	if err := c.workspaceEntity.Insert(ws); err != nil {
		return response.InternalServerError(title, "could not create the workspace")
	}
	owner := "owner"
	if err := c.ucwEntity.Insert(&identityucw.UserClientWorkspace{
		UserId: dto.Ctx.UserID, ClientId: dto.Ctx.ClientID, WorkspaceId: ws.Id, Status: &owner,
	}); err != nil {
		return response.InternalServerError(title, "could not create the workspace")
	}
	return response.Created(workspaceView(ws, owner), title, "created")
}

// UpdateMyWorkspace is PATCH /my/workspaces/:id: rename a workspace or change
// its handle. The region never changes.
func (c *AuthController) UpdateMyWorkspace(dto *UpdateMyWorkspaceDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if c.db == nil {
		return response.InternalServerError(title, "workspaces are not available")
	}
	svc := org.New(c.db.DB)
	ctx := context.Background()
	if dto.Name != nil {
		if err := svc.Rename(ctx, dto.Id, dto.Ctx.UserID, *dto.Name); err != nil {
			return orgError(title, err)
		}
	}
	if dto.Handle != nil {
		if err := svc.SetHandle(ctx, dto.Id, dto.Ctx.UserID, *dto.Handle); err != nil {
			return orgError(title, err)
		}
	}
	if dto.Name == nil && dto.Handle == nil {
		if ok, err := svc.IsAdmin(ctx, dto.Id, dto.Ctx.UserID); err != nil || !ok {
			return orgError(title, org.ErrNotAllowed)
		}
	}
	w, err := c.workspaceEntity.First(`"id" = ?`, dto.Id)
	if err != nil || w == nil {
		return response.NotFound(title, "workspace not found")
	}
	role, _ := svc.Seat(ctx, dto.Id, dto.Ctx.UserID)
	return response.Success(workspaceView(w, role), title, "saved", nil)
}
