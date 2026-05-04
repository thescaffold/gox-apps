package environment

import "github.com/thescaffold/gox-packages-core/crud"

type EnvironmentController struct {
	crud.CrudResource[Environment, CreateEnvironmentDto, UpdateEnvironmentDto]
	entity *EnvironmentEntity `inject:""`
}

func (c *EnvironmentController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Environment, CreateEnvironmentDto, UpdateEnvironmentDto]{
		Name:       "FlagEnvironment",
		Searchable: []string{"name", "workspace_id", "type_id"},
	})
}
