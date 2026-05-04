package log

import "github.com/thescaffold/gox-packages-core/crud"

type LogController struct {
	crud.CrudResource[Log, CreateLogDto, UpdateLogDto]

	entity *LogEntity `inject:""`
}

func (c *LogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Log, CreateLogDto, UpdateLogDto]{
		Name:       "CronLog",
		Searchable: []string{"job_id", "status"},
	})
}
