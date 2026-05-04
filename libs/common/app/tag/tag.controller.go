package tag

import "github.com/thescaffold/gox-packages-core/crud"

type TagController struct {
	crud.CrudResource[Tag, CreateTagDto, UpdateTagDto]
	entity *TagEntity `inject:""`
}

func (c *TagController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Tag, CreateTagDto, UpdateTagDto]{
		Name:       "CommonTag",
		Searchable: []string{"group_name", "service_name", "entity_name"},
	})
}
