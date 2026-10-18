package org

import (
	"context"
	"errors"

	"github.com/awesome-goose/goose/types"
	orgsvc "github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

type OrgController struct {
	svc *OrgService `inject:""`
}

const title = "organisation"

func fail(err error) types.Output {
	switch {
	case errors.Is(err, orgsvc.ErrNotFound):
		return response.NotFound(title, "not found")
	case errors.Is(err, orgsvc.ErrDuplicate), errors.Is(err, orgsvc.ErrHasChild), errors.Is(err, orgsvc.ErrCycle):
		return response.Conflict(title, err.Error())
	case errors.Is(err, orgsvc.ErrInvalid), errors.Is(err, orgsvc.ErrNotMember), errors.Is(err, orgsvc.ErrPermission):
		return response.BadRequest(title, err.Error())
	}
	return response.InternalServerError(title, "the change could not be made")
}

// member checks the caller holds a seat in their token's workspace.
func (c *OrgController) member(ctx ntxctx.NTXContext) (string, types.Output) {
	if ctx.UserID == "" || ctx.WorkspaceID == "" {
		return "", response.Unauthorized(title, "sign in to a workspace first")
	}
	ok, err := c.svc.Org().IsMember(context.Background(), ctx.WorkspaceID, ctx.UserID)
	if err != nil {
		return "", fail(err)
	}
	if !ok {
		return "", response.Forbidden(title, "you are not a member of this workspace")
	}
	return ctx.WorkspaceID, nil
}

// admin checks the caller is an owner or admin of their token's workspace. The
// seat is read from the database, not from the token, so a demoted admin loses
// this at once rather than when their token expires.
func (c *OrgController) admin(ctx ntxctx.NTXContext) (string, types.Output) {
	if ctx.UserID == "" || ctx.WorkspaceID == "" {
		return "", response.Unauthorized(title, "sign in to a workspace first")
	}
	ok, err := c.svc.Org().IsAdmin(context.Background(), ctx.WorkspaceID, ctx.UserID)
	if err != nil {
		return "", fail(err)
	}
	if !ok {
		return "", response.Forbidden(title, "only an owner or admin can do this")
	}
	return ctx.WorkspaceID, nil
}

func (c *OrgController) Members(dto *Ctx) types.Output {
	ws, bad := c.member(dto.Ctx)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().Members(context.Background(), ws)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "members", nil)
}

func (c *OrgController) Catalog(dto *Ctx) types.Output {
	if _, bad := c.admin(dto.Ctx); bad != nil {
		return bad
	}
	out, err := c.svc.Org().Catalog(context.Background())
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "catalog", nil)
}

// --- teams ---

func (c *OrgController) ListTeams(dto *Ctx) types.Output {
	ws, bad := c.member(dto.Ctx)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().ListTeams(context.Background(), ws)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "teams", nil)
}

func (c *OrgController) CreateTeam(dto *TeamReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	t, err := c.svc.Org().CreateTeam(context.Background(), ws, dto.Name, dto.Kind, dto.Description, dto.ParentID)
	if err != nil {
		return fail(err)
	}
	return response.Created(t, title, "created")
}

func (c *OrgController) UpdateTeam(dto *TeamUpdateReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	t, err := c.svc.Org().UpdateTeam(context.Background(), ws, dto.Id, dto.Name, dto.Description, dto.ParentID)
	if err != nil {
		return fail(err)
	}
	return response.Success(t, title, "updated", nil)
}

func (c *OrgController) DeleteTeam(dto *IDReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().DeleteTeam(context.Background(), ws, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"id": dto.Id, "deleted": true}, title, "deleted", nil)
}

func (c *OrgController) TeamMembers(dto *IDReq) types.Output {
	ws, bad := c.member(dto.Ctx)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().TeamMembers(context.Background(), ws, dto.Id)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "members", nil)
}

func (c *OrgController) AddTeamMember(dto *MemberReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().AddMember(context.Background(), ws, dto.Id, dto.UserID, dto.Role); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"teamId": dto.Id, "userId": dto.UserID}, title, "added", nil)
}

func (c *OrgController) RemoveTeamMember(dto *MemberRemoveReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().RemoveMember(context.Background(), ws, dto.Id, dto.UserID); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"teamId": dto.Id, "userId": dto.UserID, "removed": true}, title, "removed", nil)
}

// --- roles ---

func (c *OrgController) ListRoles(dto *Ctx) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().ListRoles(context.Background(), ws)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "roles", nil)
}

func (c *OrgController) CreateRole(dto *RoleReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	r, err := c.svc.Org().CreateRole(context.Background(), ws, dto.Name, dto.Description, dto.Permissions)
	if err != nil {
		return fail(err)
	}
	return response.Created(r, title, "created")
}

func (c *OrgController) UpdateRole(dto *RoleUpdateReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	r, err := c.svc.Org().UpdateRole(context.Background(), ws, dto.Id, dto.Name, dto.Description, dto.Permissions)
	if err != nil {
		return fail(err)
	}
	return response.Success(r, title, "updated", nil)
}

func (c *OrgController) DeleteRole(dto *IDReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().DeleteRole(context.Background(), ws, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"id": dto.Id, "deleted": true}, title, "deleted", nil)
}

func (c *OrgController) Assignments(dto *IDReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().Assignments(context.Background(), ws, dto.Id)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "assignments", nil)
}

func (c *OrgController) Assign(dto *AssignReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().Assign(context.Background(), ws, dto.Id, dto.SubjectType, dto.SubjectID); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"roleId": dto.Id, "subjectType": dto.SubjectType, "subjectId": dto.SubjectID}, title, "assigned", nil)
}

func (c *OrgController) Unassign(dto *UnassignReq) types.Output {
	ws, bad := c.admin(dto.Ctx)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().Unassign(context.Background(), ws, dto.Id, dto.SubjectType, dto.SubjectID); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"roleId": dto.Id, "removed": true}, title, "removed", nil)
}
