package route

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// RouteController mirrors ntx-apps/libs/controller/src/api/route/route.controller.ts:
// a vanilla CRUD controller whose uniqueness key is (group, service, type, name).
type RouteController struct {
	crud.CrudResource[Route, CreateRouteDto, UpdateRouteDto]

	entity *RouteEntity  `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *RouteController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Route, CreateRouteDto, UpdateRouteDto]{
		// Mirrors TS route.controller.ts:18 name = 'route'.
		Name: "route",
		// Mirrors TS route.controller.ts:19 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS route.controller.ts:21-23
		// unique = ({ group, service, type, name }) => [{ group, service, type, name }].
		Unique: func(d *CreateRouteDto) []utils.KeyValue {
			return []utils.KeyValue{{
				"group":   d.Group,
				"service": d.Service,
				"type":    d.Type,
				"name":    d.Name,
			}}
		},
	})
	c.SetLang(c.lang)
}
