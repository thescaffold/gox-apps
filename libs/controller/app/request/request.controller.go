package request

import "github.com/thescaffold/gox-packages/libs/core/crud"

type RequestController struct {
	crud.CrudResource[Request, CreateRequestDto, UpdateRequestDto]

	entity *RequestEntity `inject:""`
}

func (c *RequestController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Request, CreateRequestDto, UpdateRequestDto]{
		Name:       "ControllerRequest",
		Searchable: []string{"group", "service", "method", "url"},
	})
}
