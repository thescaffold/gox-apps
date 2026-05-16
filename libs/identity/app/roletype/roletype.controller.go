package roletype

import (
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// RoleTypeController mirrors ntx-apps/libs/identity/src/api/role-type/role-type.controller.ts.
// HooksCtx port (Phase S):
//   - afterCreate: seed Permission rows for the new RoleType from the DTO's
//     permissionTypeIds slice. Mirrors TS afterCreate which uses both the
//     entity (id) and the original DTO (permissionTypeIds); the legacy gox
//     HookFn doesn't surface the DTO, so this lives in HooksCtx.
//   - beforeUpdate: soft-delete any existing Permission rows linked to the
//     RoleType, then re-seed them from the supplied permissionTypeIds.
type RoleTypeController struct {
	crud.CrudResource[RoleType, CreateRoleTypeDto, UpdateRoleTypeDto]

	entity        *RoleTypeEntity                      `inject:""`
	permissionEnt *identitypermission.PermissionEntity `inject:""`
	lang          *i18n.Service                        `inject:""`
}

func (c *RoleTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[RoleType, CreateRoleTypeDto, UpdateRoleTypeDto]{
		Name:       "role group",
		Searchable: []string{"name", "desc"},
		Unique: func(d *CreateRoleTypeDto) []utils.KeyValue {
			return []utils.KeyValue{{"name": d.Name}}
		},
		HooksCtx: map[string]crud.HookCtxFn{
			crud.AfterCreate: func(event crud.HookEvent, _ ntxctx.NTXContext) error {
				rt, _ := event.Entity.(*RoleType)
				dto, _ := event.DTO.(*CreateRoleTypeDto)
				if rt == nil || dto == nil || c.permissionEnt == nil {
					return nil
				}
				for _, ptId := range dto.PermissionTypeIds {
					rtId := rt.Id
					ptCopy := ptId
					_ = c.permissionEnt.Insert(&identitypermission.Permission{
						RoleTypeId:       &rtId,
						PermissionTypeId: &ptCopy,
						Key:              ptId,
					})
				}
				return nil
			},
			crud.BeforeUpdate: func(event crud.HookEvent, _ ntxctx.NTXContext) error {
				dto, _ := event.DTO.(*UpdateRoleTypeDto)
				if dto == nil || event.ID == "" || c.permissionEnt == nil {
					return nil
				}
				if len(dto.PermissionTypeIds) == 0 {
					return nil
				}
				_, _ = c.permissionEnt.Delete(`"role_type_id" = ?`, event.ID)
				for _, ptId := range dto.PermissionTypeIds {
					rtId := event.ID
					ptCopy := ptId
					_ = c.permissionEnt.Insert(&identitypermission.Permission{
						RoleTypeId:       &rtId,
						PermissionTypeId: &ptCopy,
						Key:              ptId,
					})
				}
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}
