package flaglog

import "github.com/thescaffold/gox-packages/libs/core/crud"

type FlagLogController struct {
	crud.CrudResource[FlagLog, CreateFlagLogDto, UpdateFlagLogDto]
	entity *FlagLogEntity `inject:""`
}

func (c *FlagLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FlagLog, CreateFlagLogDto, UpdateFlagLogDto]{
		Name:       "FlagFlagLog",
		Searchable: []string{"flag_id"},
	})
}
