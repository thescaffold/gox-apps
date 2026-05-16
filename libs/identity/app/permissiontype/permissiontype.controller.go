package permissiontype

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// PermissionTypeController mirrors ntx-apps/libs/identity/src/api/permission-type/permission-type.controller.ts.
type PermissionTypeController struct {
	crud.CrudResource[PermissionType, CreatePermissionTypeDto, UpdatePermissionTypeDto]

	entity *PermissionTypeEntity `inject:""`
	lang   *i18n.Service         `inject:""`
}

func (c *PermissionTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[PermissionType, CreatePermissionTypeDto, UpdatePermissionTypeDto]{
		// Mirrors TS permission-type.controller.ts name = 'permission group'.
		Name: "permission group",
		// Mirrors TS permission-type.controller.ts searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS permission-type.controller.ts unique = ({name}) => [{name}].
		Unique: func(d *CreatePermissionTypeDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
