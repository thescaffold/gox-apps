package page

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// PageController mirrors ntx-apps/libs/blobs/src/api/page/page.controller.ts.
type PageController struct {
	crud.CrudResource[Page, CreatePageDto, UpdatePageDto]

	entity *PageEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *PageController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Page, CreatePageDto, UpdatePageDto]{
		// Mirrors TS page.controller.ts:18 name = 'page'.
		Name: "page",
		// Mirrors TS page.controller.ts:19 searchable = [].
		Searchable: nil,
		// TS unique = ({}: Page) => [] — never blocks creation.
	})
	c.SetLang(c.lang)
}
