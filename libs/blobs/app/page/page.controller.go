package page

import "github.com/thescaffold/gox-packages/libs/core/crud"

type PageController struct {
	crud.CrudResource[Page, CreatePageDto, UpdatePageDto]

	entity *PageEntity `inject:""`
}

func (c *PageController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Page, CreatePageDto, UpdatePageDto]{
		Name: "Page",
		// Mirrors TS page.controller.ts:19 searchable = [].
		Searchable: []string{},
	})
}
