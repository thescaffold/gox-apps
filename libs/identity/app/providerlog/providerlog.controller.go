package providerlog

import "github.com/thescaffold/gox-packages-core/crud"

type ProviderLogController struct {
	crud.CrudResource[ProviderLog, CreateProviderLogDto, UpdateProviderLogDto]
	entity *ProviderLogEntity `inject:""`
}

func (c *ProviderLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[ProviderLog, CreateProviderLogDto, UpdateProviderLogDto]{
		Name:       "IdentityProviderLog",
		Searchable: []string{"provider_id", "event"},
	})
}
