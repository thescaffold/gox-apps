package environmenttype

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// EnvironmentTypeController mirrors ntx-apps/libs/flags/src/api/environment-type/environment-type.controller.ts.
type EnvironmentTypeController struct {
	crud.CrudResource[EnvironmentType, CreateEnvironmentTypeDto, UpdateEnvironmentTypeDto]
	entity *EnvironmentTypeEntity `inject:""`
	lang   *i18n.Service          `inject:""`
}

func (c *EnvironmentTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[EnvironmentType, CreateEnvironmentTypeDto, UpdateEnvironmentTypeDto]{
		// Mirrors TS environment-type.controller.ts:18 name = 'environment group'.
		Name: "environment group",
		// Mirrors TS environment-type.controller.ts:19 searchable = ['name','desc','tags'].
		Searchable: []string{"name", "desc", "tags"},
		// Mirrors TS environment-type.controller.ts:21 unique = ({name}) => [{name}].
		Unique: func(d *CreateEnvironmentTypeDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
