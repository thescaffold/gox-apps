package client

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ClientController mirrors ntx-apps/libs/identity/src/api/client/client.controller.ts.
type ClientController struct {
	crud.CrudResource[Client, CreateClientDto, UpdateClientDto]

	entity *ClientEntity `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *ClientController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Client, CreateClientDto, UpdateClientDto]{
		// Mirrors TS client.controller.ts name = 'client'.
		Name: "client",
		// Mirrors TS client.controller.ts searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS client.controller.ts unique = ({key,name}) =>
		// [{name},{key}] — two OR'd constraints.
		Unique: func(d *CreateClientDto) []map[string]any {
			return []map[string]any{
				{"name": d.Name},
				{"key": d.Key},
			}
		},
	})
	c.SetLang(c.lang)
}
