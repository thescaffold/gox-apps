package sink

import "github.com/thescaffold/gox-packages/libs/core/crud"

type SinkController struct {
	crud.CrudResource[Sink, CreateSinkDto, UpdateSinkDto]
	entity *SinkEntity `inject:""`
}

func (c *SinkController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Sink, CreateSinkDto, UpdateSinkDto]{
		Name: "PolylogSink", Searchable: []string{"user_id", "workspace_id", "type_id"},
	})
}
