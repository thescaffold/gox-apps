package plantype

import (
	"sort"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// PlanTypeController mirrors ntx-apps/libs/bridge/src/api/plan-type/plan-type.controller.ts.
// Phase 7.2 wires the AfterList morph that sorts by monthly cost — TS does
// this so the cheapest plan appears first in catalog views.
type PlanTypeController struct {
	crud.CrudResource[PlanType, CreatePlanTypeDto, UpdatePlanTypeDto]

	entity *PlanTypeEntity `inject:""`
	lang   *i18n.Service   `inject:""`
}

func (c *PlanTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[PlanType, CreatePlanTypeDto, UpdatePlanTypeDto]{
		// Mirrors TS plan-type.controller.ts:19 name = 'plan group'.
		Name: "plan group",
		// Mirrors TS plan-type.controller.ts:20 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS plan-type.controller.ts:22 unique = ({licenseId,name}) =>
		// [{licenseId,name}].
		Unique: func(d *CreatePlanTypeDto) []map[string]any {
			return []map[string]any{{
				"license_id": d.LicenseId,
				"name":       d.Name,
			}}
		},
		// Mirrors TS plan-type.controller.ts:25-33 morphs.afterList: sort
		// in-place by monthly ascending.
		Morphs: map[string]crud.MorphFn{
			crud.AfterList: func(payload any, _ ntxctx.NTXContext) (any, error) {
				if list, ok := payload.([]PlanType); ok {
					sort.SliceStable(list, func(i, j int) bool {
						return list[i].Monthly < list[j].Monthly
					})
					return list, nil
				}
				return payload, nil
			},
		},
	})
	c.SetLang(c.lang)
}
