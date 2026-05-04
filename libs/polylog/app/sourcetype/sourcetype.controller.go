package sourcetype

import "github.com/thescaffold/gox-packages-core/crud"

type SourceTypeController struct {
	crud.CrudResource[SourceType, CreateSourceTypeDto, UpdateSourceTypeDto]
	entity *SourceTypeEntity `inject:""`
}

func (c *SourceTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[SourceType, CreateSourceTypeDto, UpdateSourceTypeDto]{
		Name:       "PolylogSourceType",
		Searchable: []string{"category", "name"},
	})
}
