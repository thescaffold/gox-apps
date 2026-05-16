package service

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ServiceController mirrors ntx-apps/libs/health/src/api/service/service.controller.ts.
type ServiceController struct {
	crud.CrudResource[Service, CreateServiceDto, UpdateServiceDto]

	entity *ServiceEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *ServiceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Service, CreateServiceDto, UpdateServiceDto]{
		// Mirrors TS service.controller.ts:18 name = 'service'.
		Name: "service",
		// Mirrors TS service.controller.ts:19 searchable = ['name','desc','state'].
		Searchable: []string{"name", "desc", "state"},
		// Mirrors TS service.controller.ts:21 unique = ({name}) => [{name}].
		Unique: func(d *CreateServiceDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
