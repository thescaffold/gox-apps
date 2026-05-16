package log

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// LogController mirrors ntx-apps/libs/health/src/api/log/log.controller.ts.
type LogController struct {
	crud.CrudResource[Log, CreateLogDto, UpdateLogDto]

	entity *LogEntity    `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *LogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Log, CreateLogDto, UpdateLogDto]{
		// Mirrors TS log.controller.ts:18 name = 'log'.
		Name: "log",
		// Mirrors TS log.controller.ts:19 searchable = [].
		Searchable: nil,
		// TS unique = (entity: Log) => [] — never blocks creation.
	})
	c.SetLang(c.lang)
}
