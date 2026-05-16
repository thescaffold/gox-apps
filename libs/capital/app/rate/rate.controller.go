package rate

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// RateController mirrors ntx-apps/libs/capital/src/api/rate/rate.controller.ts.
type RateController struct {
	crud.CrudResource[Rate, CreateRateDto, UpdateRateDto]

	entity *RateEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *RateController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Rate, CreateRateDto, UpdateRateDto]{
		// Mirrors TS rate.controller.ts name = 'rate'.
		Name: "rate",
		// Mirrors TS rate.controller.ts searchable = [] (empty).
		Searchable: []string{},
		// Mirrors TS rate.controller.ts unique = ({type}) => [{type}].
		Unique: func(d *CreateRateDto) []map[string]any {
			return []map[string]any{{"type": d.Type}}
		},
	})
	c.SetLang(c.lang)
}
