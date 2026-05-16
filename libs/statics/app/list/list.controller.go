package list

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ListController mirrors ntx-apps/libs/statics/src/api/list/list.controller.ts.
type ListController struct {
	crud.CrudResource[List, CreateListDto, UpdateListDto]

	entity *ListEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *ListController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[List, CreateListDto, UpdateListDto]{
		// Mirrors TS list.controller.ts:18 name = 'list'.
		Name: "list",
		// Mirrors TS list.controller.ts:19 searchable = ['value'].
		Searchable: []string{"value"},
		// Mirrors TS list.controller.ts:21 unique = ({parentId,key}) =>
		// [{parentId,key}].
		Unique: func(d *CreateListDto) []map[string]any {
			out := map[string]any{"key": d.Key}
			if d.ParentId != nil {
				out["parent_id"] = *d.ParentId
			} else {
				out["parent_id"] = nil
			}
			return []map[string]any{out}
		},
	})
	c.SetLang(c.lang)
}
