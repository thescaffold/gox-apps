package org

import (
	"context"
	"errors"
	"strings"

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
	case errors.Is(err, orgsvc.ErrNotAllowed):
		return response.Forbidden(title, err.Error())
	}
	return response.InternalServerError(title, "the change could not be made")
}

// member checks the caller holds a seat in their token's workspace.
// ctxOf is the request context carrying who is acting, for the audit trail.
func ctxOf(n ntxctx.NTXContext) context.Context {
	return orgsvc.WithActor(context.Background(), n.UserID)
}

func (c *OrgController) member(ctx ntxctx.NTXContext) (string, types.Output) {
	if ctx.UserID == "" || ctx.WorkspaceID == "" {
		return "", response.Unauthorized(title, "sign in to a workspace first")
	}
	ok, err := c.svc.Org().IsMember(ctxOf(ctx), ctx.WorkspaceID, ctx.UserID)
	if err != nil {
		return "", fail(err)
	}
	if !ok {
		return "", response.Forbidden(title, "you are not a member of this workspace")
	}
	return ctx.WorkspaceID, nil
}

// admin checks the caller may take action in their token's workspace: an owner
// or admin, unless the host's authorizer (orgsvc.SetAuthorizer) decides. The
// seat is read from the database, not from the token, so a demoted admin loses
// this at once rather than when their token expires.
func (c *OrgController) admin(ctx ntxctx.NTXContext, action string) (string, types.Output) {
	if ctx.UserID == "" || ctx.WorkspaceID == "" {
		return "", response.Unauthorized(title, "sign in to a workspace first")
	}
	ok, err := c.svc.Org().Allowed(ctxOf(ctx), ctx.WorkspaceID, ctx.UserID, action)
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
	out, err := c.svc.Org().Members(ctxOf(dto.Ctx), ws)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "members", nil)
}

func (c *OrgController) Catalog(dto *Ctx) types.Output {
	if _, bad := c.admin(dto.Ctx, orgsvc.ManageRoles); bad != nil {
		return bad
	}
	out, err := c.svc.Org().Catalog(ctxOf(dto.Ctx))
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "catalog", nil)
}

// Seats lists the roles a workspace may grant, highest first.
func (c *OrgController) Seats(dto *Ctx) types.Output {
	if _, bad := c.member(dto.Ctx); bad != nil {
		return bad
	}
	return response.Success(orgsvc.SeatRoles(), title, "seats", nil)
}

// SetSeat changes a member's seat. The service enforces who may grant what.
func (c *OrgController) SetSeat(dto *SeatReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageMembers)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().SetSeat(ctxOf(dto.Ctx), ws, dto.Ctx.UserID, dto.UserID, dto.Role); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"userId": dto.UserID, "role": strings.ToLower(strings.TrimSpace(dto.Role))}, title, "updated", nil)
}

// RemoveSeat takes a member out of the workspace.
func (c *OrgController) RemoveSeat(dto *SeatRemoveReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageMembers)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().RemoveSeat(ctxOf(dto.Ctx), ws, dto.Ctx.UserID, dto.UserID); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"userId": dto.UserID, "removed": true}, title, "removed", nil)
}

// --- teams ---

func (c *OrgController) ListTeams(dto *Ctx) types.Output {
	ws, bad := c.member(dto.Ctx)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().ListTeams(ctxOf(dto.Ctx), ws)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "teams", nil)
}

func (c *OrgController) CreateTeam(dto *TeamReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageMembers)
	if bad != nil {
		return bad
	}
	t, err := c.svc.Org().CreateTeam(ctxOf(dto.Ctx), ws, dto.Name, dto.Kind, dto.Description, dto.ParentID)
	if err != nil {
		return fail(err)
	}
	return response.Created(t, title, "created")
}

func (c *OrgController) UpdateTeam(dto *TeamUpdateReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageMembers)
	if bad != nil {
		return bad
	}
	t, err := c.svc.Org().UpdateTeam(ctxOf(dto.Ctx), ws, dto.Id, dto.Name, dto.Description, dto.ParentID)
	if err != nil {
		return fail(err)
	}
	return response.Success(t, title, "updated", nil)
}

func (c *OrgController) DeleteTeam(dto *IDReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageMembers)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().DeleteTeam(ctxOf(dto.Ctx), ws, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"id": dto.Id, "deleted": true}, title, "deleted", nil)
}

func (c *OrgController) TeamMembers(dto *IDReq) types.Output {
	ws, bad := c.member(dto.Ctx)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().TeamMembers(ctxOf(dto.Ctx), ws, dto.Id)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "members", nil)
}

func (c *OrgController) AddTeamMember(dto *MemberReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageMembers)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().AddMember(ctxOf(dto.Ctx), ws, dto.Id, dto.UserID, dto.Role); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"teamId": dto.Id, "userId": dto.UserID}, title, "added", nil)
}

func (c *OrgController) RemoveTeamMember(dto *MemberRemoveReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageMembers)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().RemoveMember(ctxOf(dto.Ctx), ws, dto.Id, dto.UserID); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"teamId": dto.Id, "userId": dto.UserID, "removed": true}, title, "removed", nil)
}

// --- roles ---

func (c *OrgController) ListRoles(dto *Ctx) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageRoles)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().ListRoles(ctxOf(dto.Ctx), ws)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "roles", nil)
}

func (c *OrgController) CreateRole(dto *RoleReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageRoles)
	if bad != nil {
		return bad
	}
	r, err := c.svc.Org().CreateRole(ctxOf(dto.Ctx), ws, dto.Name, dto.Description, dto.Permissions)
	if err != nil {
		return fail(err)
	}
	return response.Created(r, title, "created")
}

func (c *OrgController) UpdateRole(dto *RoleUpdateReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageRoles)
	if bad != nil {
		return bad
	}
	r, err := c.svc.Org().UpdateRole(ctxOf(dto.Ctx), ws, dto.Id, dto.Name, dto.Description, dto.Permissions)
	if err != nil {
		return fail(err)
	}
	return response.Success(r, title, "updated", nil)
}

func (c *OrgController) DeleteRole(dto *IDReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageRoles)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().DeleteRole(ctxOf(dto.Ctx), ws, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"id": dto.Id, "deleted": true}, title, "deleted", nil)
}

func (c *OrgController) Assignments(dto *IDReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageRoles)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().Assignments(ctxOf(dto.Ctx), ws, dto.Id)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "assignments", nil)
}

func (c *OrgController) Assign(dto *AssignReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageRoles)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().Assign(ctxOf(dto.Ctx), ws, dto.Id, dto.SubjectType, dto.SubjectID); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"roleId": dto.Id, "subjectType": dto.SubjectType, "subjectId": dto.SubjectID}, title, "assigned", nil)
}

func (c *OrgController) Unassign(dto *UnassignReq) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ManageRoles)
	if bad != nil {
		return bad
	}
	if err := c.svc.Org().Unassign(ctxOf(dto.Ctx), ws, dto.Id, dto.SubjectType, dto.SubjectID); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"roleId": dto.Id, "removed": true}, title, "removed", nil)
}

// Activity returns the workspace's recent organisation changes (admins only).
func (c *OrgController) Activity(dto *Ctx) types.Output {
	ws, bad := c.admin(dto.Ctx, orgsvc.ReadAudit)
	if bad != nil {
		return bad
	}
	out, err := c.svc.Org().Activity(ctxOf(dto.Ctx), ws, 100)
	if err != nil {
		return fail(err)
	}
	return response.Success(out, title, "activity", nil)
}
