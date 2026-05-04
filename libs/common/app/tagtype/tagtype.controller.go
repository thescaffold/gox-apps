package tagtype

import "github.com/thescaffold/gox-packages-core/crud"

type TagTypeController struct {
	crud.CrudResource[TagType, CreateTagTypeDto, UpdateTagTypeDto]
	entity *TagTypeEntity `inject:""`
}

func (c *TagTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[TagType, CreateTagTypeDto, UpdateTagTypeDto]{
		Name:       "CommonTagType",
		Searchable: []string{"name"},
	})
}
