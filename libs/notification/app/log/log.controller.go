package log

import "github.com/thescaffold/gox-packages/libs/core/crud"

type LogController struct {
	crud.CrudResource[Log, CreateLogDto, UpdateLogDto]
	entity *LogEntity `inject:""`
}

func (c *LogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Log, CreateLogDto, UpdateLogDto]{
		Name:       "NotificationLog",
		Searchable: []string{"reference", "user_id", "workspace_id", "key"},
	})
}
