package provider

import "github.com/thescaffold/gox-packages-core/crud"

type ProviderController struct {
	crud.CrudResource[Provider, CreateProviderDto, UpdateProviderDto]
	entity *ProviderEntity `inject:""`
}

func (c *ProviderController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Provider, CreateProviderDto, UpdateProviderDto]{
		Name:       "IdentityProvider",
		Searchable: []string{"user_id", "type", "reference"},
	})
}
