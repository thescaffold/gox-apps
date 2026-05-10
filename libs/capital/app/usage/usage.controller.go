package usage

import "github.com/thescaffold/gox-packages/libs/core/crud"

type UsageController struct {
	crud.CrudResource[Usage, CreateUsageDto, UpdateUsageDto]
	entity *UsageEntity `inject:""`
}

func (c *UsageController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Usage, CreateUsageDto, UpdateUsageDto]{
		Name:       "CapitalUsage",
		Searchable: []string{"user_id", "client_id", "workspace_id", "rate_id"},
	})
}
