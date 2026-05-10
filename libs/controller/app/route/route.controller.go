package route

import "github.com/thescaffold/gox-packages/libs/core/crud"

type RouteController struct {
	crud.CrudResource[Route, CreateRouteDto, UpdateRouteDto]

	entity *RouteEntity `inject:""`
}

func (c *RouteController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Route, CreateRouteDto, UpdateRouteDto]{
		Name: "route",
		// Mirrors TS route.controller.ts:19 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
	})
}
