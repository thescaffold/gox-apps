package clientlog

import "github.com/thescaffold/gox-packages/libs/core/crud"

type ClientLogController struct {
	crud.CrudResource[ClientLog, CreateClientLogDto, UpdateClientLogDto]
	entity *ClientLogEntity `inject:""`
}

func (c *ClientLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[ClientLog, CreateClientLogDto, UpdateClientLogDto]{
		Name:       "IdentityClientLog",
		Searchable: []string{"client_id", "event"},
	})
}
