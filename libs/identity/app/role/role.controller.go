package role

import (
	"strings"

	"github.com/awesome-goose/goose/types"
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
