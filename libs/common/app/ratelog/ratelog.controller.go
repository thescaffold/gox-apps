package ratelog

import "github.com/thescaffold/gox-packages/libs/core/crud"

type RateLogController struct {
	crud.CrudResource[RateLog, CreateRateLogDto, UpdateRateLogDto]
	entity *RateLogEntity `inject:""`
}

func (c *RateLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[RateLog, CreateRateLogDto, UpdateRateLogDto]{
		Name:       "CommonRateLog",
		Searchable: []string{"rate_id"},
	})
}
