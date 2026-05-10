package client

import "github.com/thescaffold/gox-packages/libs/core/crud"

type ClientController struct {
	crud.CrudResource[Client, CreateClientDto, UpdateClientDto]
	entity *ClientEntity `inject:""`
}

func (c *ClientController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Client, CreateClientDto, UpdateClientDto]{
		Name:       "IdentityClient",
		Searchable: []string{"key", "name", "type"},
	})
}
