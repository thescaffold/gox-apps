package plantype

import (
	"sort"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// PlanTypeController mirrors ntx-apps/libs/capital/src/api/plan-type/plan-type.controller.ts.
// Implements:
//   - morphs.afterGet: adjust daily/weekly/monthly/yearly + prorated costs
//     via CalculateAdaptiveCost / CalculateProratedCost; nil-return when the
//     entity's currency doesn't match the preference currency (TS-equivalent
//     visibility filter).
//   - morphs.afterList: same adjust + sort ascending by monthly.
type PlanTypeController struct {
	crud.CrudResource[PlanType, CreatePlanTypeDto, UpdatePlanTypeDto]

	entity *PlanTypeEntity `inject:""`
	lang   *i18n.Service   `inject:""`
}

var allowedCurrencies = map[string]bool{"USD": true, "NGN": true}

const defaultCurrency = "USD"

func (c *PlanTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[PlanType, CreatePlanTypeDto, UpdatePlanTypeDto]{
		Name:       "plan group",
		Searchable: []string{"name", "desc"},
		Unique: func(d *CreatePlanTypeDto) []utils.KeyValue {
			return []utils.KeyValue{{"client_id": d.ClientId, "name": d.Name}}
		},
		Morphs: map[string]crud.MorphFn{
			crud.AfterGet: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				pt, ok := payload.(*PlanType)
				if !ok || pt == nil {
					return payload, nil
				}
				if cur := preferenceCurrency(ctx); cur != "" && pt.Currency != cur {
					return nil, nil
				}
				return adjustCost(pt), nil
			},
			crud.AfterList: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				rows, ok := payload.([]PlanType)
				if !ok {
					return payload, nil
				}
				out := make([]map[string]any, 0, len(rows))
				for i := range rows {
					out = append(out, adjustCost(&rows[i]))
				}
				sort.SliceStable(out, func(i, j int) bool {
					return intOf(out[i]["monthly"]) < intOf(out[j]["monthly"])
				})
				return out, nil
			},
		},
	})
	c.SetLang(c.lang)
}

// adjustCost mirrors TS PlanTypeController.adjustCost — flat cost + prorated
// sub-object. rate is fixed to 1.0 in the TS impl (the rate parameter is
// preserved for future PlatformService.statics integration).
func adjustCost(pt *PlanType) map[string]any {
	out := map[string]any{
		"id":        pt.Id,
		"clientId":  pt.ClientId,
		"key":       pt.Key,
		"name":      pt.Name,
		"desc":      pt.Desc,
		"detail":    pt.Detail,
		"type":      pt.Type,
		"currency":  pt.Currency,
		"meta":      pt.Meta,
		"status":    pt.Status,
		"createdAt": pt.CreatedAt,
		"updatedAt": pt.UpdatedAt,
		"daily":     utils.CalculateAdaptiveCost(floatOf(pt.Daily), 1),
		"weekly":    utils.CalculateAdaptiveCost(floatOf(pt.Weekly), 1),
		"monthly":   utils.CalculateAdaptiveCost(float64(pt.Monthly), 1),
		"yearly":    utils.CalculateAdaptiveCost(floatOf(pt.Yearly), 1),
		"prorated": map[string]any{
			"daily":   utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(floatOf(pt.Daily))), 1),
			"weekly":  utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(floatOf(pt.Weekly))), 1),
			"monthly": utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(float64(pt.Monthly))), 1),
			"yearly":  utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(floatOf(pt.Yearly))), 1),
		},
	}
	return out
}

func floatOf(p *int) float64 {
	if p == nil {
		return 0
	}
	return float64(*p)
}

func intOf(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	}
	return 0
}

func preferenceCurrency(ctx ntxctx.NTXContext) string {
	if ctx.Preference == nil {
		return ""
	}
	if c, ok := ctx.Preference["currency"].(string); ok {
		return c
	}
	return ""
}
