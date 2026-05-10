package sinktype

import "github.com/thescaffold/gox-packages-core/crud"

type SinkTypeController struct {
	crud.CrudResource[SinkType, CreateSinkTypeDto, UpdateSinkTypeDto]
	entity *SinkTypeEntity `inject:""`
}

func (c *SinkTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[SinkType, CreateSinkTypeDto, UpdateSinkTypeDto]{
		Name: "PolylogSinkType", Searchable: []string{"category", "name"},
	})
}
