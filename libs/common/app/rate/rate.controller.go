package rate

import "github.com/thescaffold/gox-packages-core/crud"

type RateController struct {
	crud.CrudResource[Rate, CreateRateDto, UpdateRateDto]
	entity *RateEntity `inject:""`
}

func (c *RateController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Rate, CreateRateDto, UpdateRateDto]{
		Name:       "CommonRate",
		Searchable: []string{"currency"},
	})
}
