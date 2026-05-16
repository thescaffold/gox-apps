package list

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ListController mirrors ntx-apps/libs/cache/src/api/list/list.controller.ts.
type ListController struct {
	crud.CrudResource[List, CreateListDto, UpdateListDto]

	entity *ListEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *ListController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[List, CreateListDto, UpdateListDto]{
		// Mirrors TS list.controller.ts:18 name = 'list'.
		Name: "list",
		// Mirrors TS list.controller.ts:19 searchable = ['key','value'].
		Searchable: []string{"key", "value"},
		// Mirrors TS list.controller.ts:21 unique = ({ key }) => [{ key }].
		Unique: func(d *CreateListDto) []map[string]any {
			return []map[string]any{{"key": d.Key}}
		},
	})
	c.SetLang(c.lang)
}
