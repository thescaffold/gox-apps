package history

import "github.com/thescaffold/gox-packages-core/crud"

type HistoryController struct {
	crud.CrudResource[History, CreateHistoryDto, UpdateHistoryDto]

	entity *HistoryEntity `inject:""`
}

func (c *HistoryController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[History, CreateHistoryDto, UpdateHistoryDto]{
		Name: "history",
		// Mirrors TS history.controller.ts:20 searchable = ['service'].
		// Note: Go History entity has no `service` column; keeping local-only
		// fields means search degrades silently. This intentionally preserves
		// the TS contract — controller-side searchable list.
		Searchable: []string{"service"},
	})
}
