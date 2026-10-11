package auth

import (
	"strings"

	"github.com/awesome-goose/goose/types"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
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

// CreateMyWorkspaceDto carries POST /my/workspaces.
type CreateMyWorkspaceDto struct {
	Ctx  ntxctx.NTXContext `context:"ntx"`
	Name string            `json:"name" binding:"required"`
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
		out = append(out, map[string]any{"id": w.Id, "name": w.Name, "role": role})
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
	if err := c.workspaceEntity.Insert(ws); err != nil {
		return response.InternalServerError(title, "could not create the workspace")
	}
	owner := "owner"
	if err := c.ucwEntity.Insert(&identityucw.UserClientWorkspace{
		UserId: dto.Ctx.UserID, ClientId: dto.Ctx.ClientID, WorkspaceId: ws.Id, Status: &owner,
	}); err != nil {
		return response.InternalServerError(title, "could not create the workspace")
	}
	return response.Created(map[string]any{"id": ws.Id, "name": ws.Name, "role": owner}, title, "created")
}
