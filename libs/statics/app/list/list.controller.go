package list

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
)

type ListController struct {
	crud.CrudResource[List, CreateListDto, UpdateListDto]

	entity *ListEntity `inject:""`
}

func (c *ListController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[List, CreateListDto, UpdateListDto]{
		Name: "List",
		// Mirrors TS list.controller.ts:19 searchable = ['value'].
		Searchable: []string{"value"},
	})
}
