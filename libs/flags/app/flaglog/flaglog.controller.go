package flaglog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FlagLogController mirrors ntx-apps/libs/flags/src/api/flag-log/flag-log.controller.ts.
type FlagLogController struct {
	crud.CrudResource[FlagLog, CreateFlagLogDto, UpdateFlagLogDto]
	entity *FlagLogEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *FlagLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FlagLog, CreateFlagLogDto, UpdateFlagLogDto]{
		// Mirrors TS flag-log.controller.ts:18 name = 'flag log'.
		Name: "flag log",
		// Mirrors TS flag-log.controller.ts:19 searchable = [].
		Searchable: nil,
		// TS unique = (entity: FlagLog) => [] — never blocks creation.
	})
	c.SetLang(c.lang)
}
