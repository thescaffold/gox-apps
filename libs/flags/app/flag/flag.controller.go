package flag

import "github.com/thescaffold/gox-packages/libs/core/crud"

type FlagController struct {
	crud.CrudResource[Flag, CreateFlagDto, UpdateFlagDto]
	entity *FlagEntity `inject:""`
}

func (c *FlagController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Flag, CreateFlagDto, UpdateFlagDto]{
		Name:       "FlagFlag",
		Searchable: []string{"name", "workspace_id", "environment_id", "level"},
	})
}
