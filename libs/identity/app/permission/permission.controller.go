package permission

import (
	"strings"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// PermissionController mirrors ntx-apps/libs/identity/src/api/permission/permission.controller.ts.
// Ported behaviour: standard CRUD + Attach/Detach/Sync that bind a roleType
// to a set of permissionTypes. Uses the Phase H shadow columns (role_type_id,
// permission_type_id) added by AlignIdentityWithNtx.
type PermissionController struct {
	crud.CrudResource[Permission, CreatePermissionDto, UpdatePermissionDto]

	entity *PermissionEntity `inject:""`
	lang   *i18n.Service     `inject:""`
}

func (c *PermissionController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Permission, CreatePermissionDto, UpdatePermissionDto]{
		Name:       "permission",
		Searchable: []string{},
		// TS unique = ({roleTypeId,permissionTypeId}) — Phase H wires the
		// shadow columns; legacy (role_id, key) remains as a secondary
		// constraint until cleanup migration retires it.
		Unique: func(d *CreatePermissionDto) []utils.KeyValue {
			out := []utils.KeyValue{}
			if d.RoleTypeId != nil && d.PermissionTypeId != nil {
				out = append(out, utils.KeyValue{
					"role_type_id":       *d.RoleTypeId,
					"permission_type_id": *d.PermissionTypeId,
				})
			}
			legacy := utils.KeyValue{"key": d.Key}
			if d.RoleId != nil {
				legacy["role_id"] = *d.RoleId
			} else {
				legacy["role_id"] = nil
			}
			out = append(out, legacy)
			return out
		},
	})
	c.SetLang(c.lang)
}

// Attach mirrors TS @Post('attach') — bulk-insert one Permission row per
// supplied permissionTypeId, all bound to the same roleTypeId.
func (c *PermissionController) Attach(dto *LinkPermissionDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.permission.title", nil, pref)
	out := make([]Permission, 0, len(dto.PermissionTypeIds))
	for _, pt := range dto.PermissionTypeIds {
		ptCopy := pt
		rtCopy := dto.RoleTypeId
		row := &Permission{
			RoleTypeId:       &rtCopy,
			PermissionTypeId: &ptCopy,
			Key:              pt,
			Scope:            "workspace",
			Resource:         "*",
			Action:           "*",
		}
		if err := c.entity.Insert(row); err == nil {
			out = append(out, *row)
		}
	}
	return response.Success(out, title,
		c.lang.Translate("apps.identity.permission.post.attach.success", nil, pref), nil)
}

// Detach mirrors TS @Post('detach') — soft-delete rows matching
// (roleTypeId, permissionTypeId IN (...)).
func (c *PermissionController) Detach(dto *LinkPermissionDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.permission.title", nil, pref)
	placeholders, args := inClause(dto.PermissionTypeIds)
	where := `"role_type_id" = ? AND "permission_type_id" IN ` + placeholders
	allArgs := append([]any{dto.RoleTypeId}, args...)
	_, _ = c.entity.Delete(where, allArgs...)
	return response.Success(dto, title,
		c.lang.Translate("apps.identity.permission.post.detach.success", nil, pref), nil)
}

// Sync mirrors TS @Post('sync') — wipe all rows for the roleTypeId, then
// attach the supplied permissionTypeIds. Used to replace the role's full
// permission set in one shot.
func (c *PermissionController) Sync(dto *LinkPermissionDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.permission.title", nil, pref)
	_, _ = c.entity.Delete(`"role_type_id" = ?`, dto.RoleTypeId)
	out := make([]Permission, 0, len(dto.PermissionTypeIds))
	for _, pt := range dto.PermissionTypeIds {
		ptCopy := pt
		rtCopy := dto.RoleTypeId
		row := &Permission{
			RoleTypeId:       &rtCopy,
			PermissionTypeId: &ptCopy,
			Key:              pt,
			Scope:            "workspace",
			Resource:         "*",
			Action:           "*",
		}
		if err := c.entity.Insert(row); err == nil {
			out = append(out, *row)
		}
	}
	return response.Success(out, title,
		c.lang.Translate("apps.identity.permission.post.sync.success", nil, pref), nil)
}

func inClause(ids []string) (string, []any) {
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	return "(" + strings.Join(placeholders, ",") + ")", args
}
