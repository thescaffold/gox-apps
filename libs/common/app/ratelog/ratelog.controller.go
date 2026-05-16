package ratelog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// RateLogController mirrors ntx-apps/libs/common/src/api/rate-log/rate-log.controller.ts.
type RateLogController struct {
	crud.CrudResource[RateLog, CreateRateLogDto, UpdateRateLogDto]

	entity *RateLogEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *RateLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[RateLog, CreateRateLogDto, UpdateRateLogDto]{
		// Mirrors TS rate-log.controller.ts:18 name = 'rate log'.
		Name: "rate log",
		// Mirrors TS rate-log.controller.ts:19 searchable = [] (empty).
		Searchable: []string{},
		// TS unique = () => [] (no uniqueness constraints) — left unset.
	})
	c.SetLang(c.lang)
}
