package providerlog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ProviderLogController mirrors ntx-apps/libs/identity/src/api/provider-log/provider-log.controller.ts.
type ProviderLogController struct {
	crud.CrudResource[ProviderLog, CreateProviderLogDto, UpdateProviderLogDto]

	entity *ProviderLogEntity `inject:""`
	lang   *i18n.Service      `inject:""`
}

func (c *ProviderLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[ProviderLog, CreateProviderLogDto, UpdateProviderLogDto]{
		// Mirrors TS provider-log.controller.ts name = 'provider log'.
		Name: "provider log",
		// Mirrors TS provider-log.controller.ts searchable = [] (empty).
		Searchable: []string{},
		// TS unique = () => [] — no uniqueness constraints; left unset.
	})
	c.SetLang(c.lang)
}
