package list

import "github.com/thescaffold/gox-packages-core/crud"

type ListController struct {
	crud.CrudResource[List, CreateListDto, UpdateListDto]

	entity *ListEntity `inject:""`
}

func (c *ListController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[List, CreateListDto, UpdateListDto]{
		Name:       "CacheList",
		Searchable: []string{"key", "group"},
	})
}
