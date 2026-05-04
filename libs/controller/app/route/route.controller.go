package route

import "github.com/thescaffold/gox-packages-core/crud"

type RouteController struct {
	crud.CrudResource[Route, CreateRouteDto, UpdateRouteDto]

	entity *RouteEntity `inject:""`
}

func (c *RouteController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Route, CreateRouteDto, UpdateRouteDto]{
		Name:       "ControllerRoute",
		Searchable: []string{"group", "service", "name", "type"},
	})
}
