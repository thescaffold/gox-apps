package environmenttype

import "github.com/thescaffold/gox-packages/libs/core/crud"

type EnvironmentTypeController struct {
	crud.CrudResource[EnvironmentType, CreateEnvironmentTypeDto, UpdateEnvironmentTypeDto]
	entity *EnvironmentTypeEntity `inject:""`
}

func (c *EnvironmentTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[EnvironmentType, CreateEnvironmentTypeDto, UpdateEnvironmentTypeDto]{
		Name:       "FlagEnvironmentType",
		Searchable: []string{"category", "name"},
	})
}
