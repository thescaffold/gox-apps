package role

import (
	"net/http"
	"strings"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/rbac"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// RoleController mirrors ntx-apps/libs/identity/src/api/role/role.controller.ts.
// Ported behaviour:
//   - morphs.beforeCreate: hydrate clientId/workspaceId from context.
//   - Attach: save one Role row per supplied roleTypeId, keyed by
//     (entityName, entityId, clientId, workspaceId).
//   - Detach: soft-delete rows matching (entityName, entityId, clientId,
//     workspaceId, roleTypeId IN (...)).
//   - Sync: full replace — soft-delete all rows for the entity, then attach.
type RoleController struct {
	crud.CrudResource[Role, CreateRoleDto, UpdateRoleDto]

	entity *RoleEntity   `inject:""`
	db     *sql.Db       `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *RoleController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Role, CreateRoleDto, UpdateRoleDto]{
		Name:       "role",
		Searchable: []string{},
		Unique: func(d *CreateRoleDto) []map[string]any {
			out := map[string]any{"name": d.Name}
			if d.WorkspaceId != nil {
				out["workspace_id"] = *d.WorkspaceId
			} else {
				out["workspace_id"] = nil
			}
			return []map[string]any{out}
		},
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateRoleDto)
				if !ok || p == nil {
					return nil, nil
				}
				if err := c.refusePlatformRole(ctx, p.Name, p.RoleTypeId); err != nil {
					return nil, err
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = &id
					}
				}
				return p, nil
			},
			crud.BeforeUpdate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*UpdateRoleDto)
				if !ok || p == nil {
					return nil, nil
				}
				name := ""
				if p.Name != nil {
					name = *p.Name
				}
				return nil, c.refusePlatformRole(ctx, name, p.RoleTypeId)
			},
		},
	})
	c.SetLang(c.lang)
}

// Attach mirrors TS @Post('attach') — bulk-insert one Role row per supplied
// roleTypeId, all keyed by the (entityName, entityId, clientId, workspaceId)
// envelope.
func (c *RoleController) Attach(dto *LinkRoleDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.role.title", nil, pref)
	if reason := c.scopeLink(dto); reason != "" {
		return response.Forbidden(title, reason)
	}
	out := make([]Role, 0, len(dto.RoleTypeIds))
	for _, rt := range dto.RoleTypeIds {
		rtCopy := rt
		row := &Role{
			Name:        rt,
			ClientId:    dto.ClientId,
			WorkspaceId: dto.WorkspaceId,
			Type:        dto.EntityName,
			EntityName:  &dto.EntityName,
			EntityId:    &dto.EntityId,
			RoleTypeId:  &rtCopy,
		}
		if err := c.entity.Insert(row); err == nil {
			out = append(out, *row)
		}
	}
	return response.Success(out, title,
		c.lang.Translate("apps.identity.role.post.attach.success", nil, pref), nil)
}

// Detach mirrors TS @Post('detach') — soft-delete rows matching the supplied
// (entityName, entityId, clientId, workspaceId, roleTypeId IN (...)).
func (c *RoleController) Detach(dto *LinkRoleDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.role.title", nil, pref)
	if reason := c.scopeLink(dto); reason != "" {
		return response.Forbidden(title, reason)
	}
	placeholders, args := inClause(dto.RoleTypeIds)
	where := `"entity_name" = ? AND "entity_id" = ? AND "client_id" = ? AND "role_type_id" IN ` + placeholders
	allArgs := append([]any{dto.EntityName, dto.EntityId, dto.ClientId}, args...)
	if dto.WorkspaceId != nil {
		where += ` AND "workspace_id" = ?`
		allArgs = append(allArgs, *dto.WorkspaceId)
	}
	_, _ = c.entity.Delete(where, allArgs...)
	return response.Success(dto, title,
		c.lang.Translate("apps.identity.role.post.detach.success", nil, pref), nil)
}

// Sync mirrors TS @Post('sync') — wipe all rows for the entity, then attach
// the supplied roleTypeIds. Used to replace the entity's full role-set in
// one shot.
func (c *RoleController) Sync(dto *LinkRoleDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.role.title", nil, pref)
	if reason := c.scopeLink(dto); reason != "" {
		return response.Forbidden(title, reason)
	}
	where := `"entity_name" = ? AND "entity_id" = ? AND "client_id" = ?`
	args := []any{dto.EntityName, dto.EntityId, dto.ClientId}
	if dto.WorkspaceId != nil {
		where += ` AND "workspace_id" = ?`
		args = append(args, *dto.WorkspaceId)
	}
	_, _ = c.entity.Delete(where, args...)
	out := make([]Role, 0, len(dto.RoleTypeIds))
	for _, rt := range dto.RoleTypeIds {
		rtCopy := rt
		row := &Role{
			Name:        rt,
			ClientId:    dto.ClientId,
			WorkspaceId: dto.WorkspaceId,
			Type:        dto.EntityName,
			EntityName:  &dto.EntityName,
			EntityId:    &dto.EntityId,
			RoleTypeId:  &rtCopy,
		}
		if err := c.entity.Insert(row); err == nil {
			out = append(out, *row)
		}
	}
	return response.Success(out, title,
		c.lang.Translate("apps.identity.role.post.sync.success", nil, pref), nil)
}

// scopeLink confines a workspace admin's link to their own client and workspace
// and away from platform roles, filling the scope in when the body leaves it
// out. It returns why a link is refused, or "". rbac.Guard has already decided
// the caller may write at all.
func (c *RoleController) scopeLink(dto *LinkRoleDto) string {
	if rbac.PlatformAdmin(dto.Ctx) {
		return ""
	}
	if dto.ClientId != "" && dto.ClientId != dto.Ctx.ClientID {
		return "you may only manage roles in your own client"
	}
	if dto.WorkspaceId != nil && *dto.WorkspaceId != dto.Ctx.WorkspaceID {
		return "you may only manage roles in your own workspace"
	}
	workspace := dto.Ctx.WorkspaceID
	dto.ClientId, dto.WorkspaceId = dto.Ctx.ClientID, &workspace
	if c.anyPlatformRole(dto.RoleTypeIds) {
		return "you may not manage platform roles"
	}
	return ""
}

// anyPlatformRole reports whether an id, or the role type it names, is a
// platform role. An unreadable role type counts as one.
func (c *RoleController) anyPlatformRole(ids []string) bool {
	for _, id := range ids {
		if rbac.Privileged(id) {
			return true
		}
	}
	if len(ids) == 0 {
		return false
	}
	var names []string
	if err := c.db.Raw(`SELECT "name" FROM "IdentityRoleTypes" WHERE "id" IN ?`, ids).Scan(&names).Error; err != nil {
		return true
	}
	for _, name := range names {
		if rbac.Privileged(name) {
			return true
		}
	}
	return false
}

// refusePlatformRole stops a direct row write naming a platform role.
func (c *RoleController) refusePlatformRole(ctx ntxctx.NTXContext, name string, roleTypeId *string) error {
	if rbac.PlatformAdmin(ctx) {
		return nil
	}
	ids := []string{}
	if name != "" {
		ids = append(ids, name)
	}
	if roleTypeId != nil {
		ids = append(ids, *roleTypeId)
	}
	if c.anyPlatformRole(ids) {
		return &crud.HTTPError{Status: http.StatusForbidden, Message: "you may not manage platform roles"}
	}
	return nil
}

// inClause builds the `(?, ?, ?)` placeholder + args slice for an IN clause.
func inClause(ids []string) (string, []any) {
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	return "(" + strings.Join(placeholders, ",") + ")", args
}
