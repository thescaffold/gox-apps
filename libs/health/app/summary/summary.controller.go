package summary

import "github.com/thescaffold/gox-packages-core/crud"

type SummaryController struct {
	crud.CrudResource[Summary, CreateSummaryDto, UpdateSummaryDto]

	entity *SummaryEntity `inject:""`
}

func (c *SummaryController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Summary, CreateSummaryDto, UpdateSummaryDto]{
		Name:       "HealthSummary",
		Searchable: []string{"service_id", "type"},
	})
}
