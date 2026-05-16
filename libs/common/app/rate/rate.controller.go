package rate

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// RateController mirrors ntx-apps/libs/common/src/api/rate/rate.controller.ts.
type RateController struct {
	crud.CrudResource[Rate, CreateRateDto, UpdateRateDto]

	entity *RateEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *RateController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Rate, CreateRateDto, UpdateRateDto]{
		// Mirrors TS rate.controller.ts:18 name = 'rate'.
		Name: "rate",
		// Mirrors TS rate.controller.ts:19 searchable = ['currency'].
		Searchable: []string{"currency"},
		// TS unique = () => [] (no uniqueness constraints) — left unset here.
	})
	c.SetLang(c.lang)
}
