package service

import "github.com/thescaffold/gox-packages-core/crud"

type ServiceController struct {
	crud.CrudResource[Service, CreateServiceDto, UpdateServiceDto]

	entity *ServiceEntity `inject:""`
}

func (c *ServiceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Service, CreateServiceDto, UpdateServiceDto]{
		Name:       "HealthService",
		Searchable: []string{"name", "type", "state"},
	})
}
