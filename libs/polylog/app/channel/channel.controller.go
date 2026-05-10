package channel

import "github.com/thescaffold/gox-packages-core/crud"

type ChannelController struct {
	crud.CrudResource[Channel, CreateChannelDto, UpdateChannelDto]
	entity *ChannelEntity `inject:""`
}

func (c *ChannelController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Channel, CreateChannelDto, UpdateChannelDto]{
		Name: "PolylogChannel", Searchable: []string{"user_id", "workspace_id", "source_id"},
	})
}
