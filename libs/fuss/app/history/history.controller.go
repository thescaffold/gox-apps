package history

import "github.com/thescaffold/gox-packages-core/crud"

type HistoryController struct {
	crud.CrudResource[History, CreateHistoryDto, UpdateHistoryDto]

	entity *HistoryEntity `inject:""`
}

func (c *HistoryController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[History, CreateHistoryDto, UpdateHistoryDto]{
		Name:       "FussHistory",
		Searchable: []string{"user_id", "query", "type"},
	})
}
