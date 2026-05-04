package source

import "github.com/thescaffold/gox-packages-core/crud"

type SourceController struct {
	crud.CrudResource[Source, CreateSourceDto, UpdateSourceDto]
	entity *SourceEntity `inject:""`
}

func (c *SourceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Source, CreateSourceDto, UpdateSourceDto]{
		Name:       "PolylogSource",
		Searchable: []string{"user_id", "workspace_id", "category", "type_id"},
	})
}
