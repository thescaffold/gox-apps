package clientlog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ClientLogController mirrors ntx-apps/libs/identity/src/api/client-log/client-log.controller.ts.
type ClientLogController struct {
	crud.CrudResource[ClientLog, CreateClientLogDto, UpdateClientLogDto]

	entity *ClientLogEntity `inject:""`
	lang   *i18n.Service    `inject:""`
}

func (c *ClientLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[ClientLog, CreateClientLogDto, UpdateClientLogDto]{
		// Mirrors TS client-log.controller.ts name = 'client log'.
		Name: "client log",
		// Mirrors TS client-log.controller.ts searchable = [] (empty).
		Searchable: []string{},
		// TS unique = () => [] — no uniqueness constraints; left unset.
	})
	c.SetLang(c.lang)
}
