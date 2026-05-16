package summary

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// SummaryController mirrors ntx-apps/libs/health/src/api/summary/summary.controller.ts.
type SummaryController struct {
	crud.CrudResource[Summary, CreateSummaryDto, UpdateSummaryDto]

	entity *SummaryEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *SummaryController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Summary, CreateSummaryDto, UpdateSummaryDto]{
		// Mirrors TS summary.controller.ts:18 name = 'summary'.
		Name: "summary",
		// Mirrors TS summary.controller.ts:19 searchable = ['note'].
		Searchable: []string{"note"},
		// Mirrors TS summary.controller.ts:21
		// unique = ({serviceId, type}) => [{serviceId, type}].
		Unique: func(d *CreateSummaryDto) []map[string]any {
			return []map[string]any{{"service_id": d.ServiceId, "type": d.Type}}
		},
	})
	c.SetLang(c.lang)
}
