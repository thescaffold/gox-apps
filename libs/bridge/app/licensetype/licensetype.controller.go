package licensetype

import (
	"encoding/json"

	staticsapp "github.com/thescaffold/gox-apps/libs/statics/app"
	staticslist "github.com/thescaffold/gox-apps/libs/statics/app/list"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// LicenseTypeController mirrors ntx-apps/libs/bridge/src/api/license-type/license-type.controller.ts.
// Phase 7.2 wires the AfterGet/AfterList cost-adjustment morphs that take the
// price columns (daily/weekly/monthly/yearly), look up the request currency's
// FX rate via StaticsAppService, and rewrite the prices using
// CalculateAdaptiveCost (+ a prorated variant). TS uses
// PlatformService.apps.statics.get.code(currency, country) — gox short-circuits
// that remote call by reading the statics List entity directly through the
// in-process StaticsAppService injection.
type LicenseTypeController struct {
	crud.CrudResource[LicenseType, CreateLicenseTypeDto, UpdateLicenseTypeDto]

	entity     *LicenseTypeEntity     `inject:""`
	staticsApp *staticsapp.AppService `inject:""`
	lang       *i18n.Service          `inject:""`
}

func (c *LicenseTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[LicenseType, CreateLicenseTypeDto, UpdateLicenseTypeDto]{
		// Mirrors TS license-type.controller.ts:25 name = 'license group'.
		Name: "license group",
		// Mirrors TS license-type.controller.ts:26 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS license-type.controller.ts:28 unique = ({key,name}) =>
		// [{name},{key}] — two OR'd constraints.
		Unique: func(d *CreateLicenseTypeDto) []map[string]any {
			return []map[string]any{
				{"name": d.Name},
				{"key": d.Key},
			}
		},
		// Mirrors TS license-type.controller.ts:30-58 morphs.afterGet/afterList:
		// scale prices into the preference currency and add a prorated block.
		Morphs: map[string]crud.MorphFn{
			crud.AfterGet: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				if lt, ok := payload.(*LicenseType); ok && lt != nil {
					return c.adjust(lt, ctx), nil
				}
				return payload, nil
			},
			crud.AfterList: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				if list, ok := payload.([]LicenseType); ok {
					out := make([]map[string]any, 0, len(list))
					for i := range list {
						out = append(out, c.adjust(&list[i], ctx))
					}
					// TS additionally sorts by monthly cost; we sort in-place
					// using the adjusted monthly value below.
					sortByMonthly(out)
					return out, nil
				}
				return payload, nil
			},
		},
	})
	c.SetLang(c.lang)
}

// adjust scales the license-type's daily/weekly/monthly/yearly using the
// statics currency rate, mirroring TS adjustCost(). Returns a map so the
// shape can carry the original entity fields plus a `prorated` sub-object
// and the resolved currency metadata, exactly as TS does.
func (c *LicenseTypeController) adjust(lt *LicenseType, ctx ntxctx.NTXContext) map[string]any {
	// TS: const { currency, country } = preference; const meta = await
	// platformService.apps.statics.get.code(currency, country).meta;
	// gox: read the statics entry directly via the injected AppService.
	currency := ""
	country := ""
	if ctx.Preference != nil {
		if v, ok := ctx.Preference["currency"].(string); ok {
			currency = v
		}
		if v, ok := ctx.Preference["country"].(string); ok {
			country = v
		}
	}

	rate := 1.0
	if c.staticsApp != nil && currency != "" {
		var stat any
		if country != "" {
			parent := country
			stat, _ = c.staticsApp.FindByCode(currency, &parent)
		}
		if stat == nil {
			stat, _ = c.staticsApp.FindByCode(currency, nil)
		}
		if l, ok := stat.(*staticslist.List); ok && l != nil {
			rate = parseRate(l.Meta, rate)
		}
	}

	d := utils.CalculateAdaptiveCost(deref(lt.Daily), rate)
	w := utils.CalculateAdaptiveCost(deref(lt.Weekly), rate)
	m := utils.CalculateAdaptiveCost(lt.Monthly, rate)
	y := utils.CalculateAdaptiveCost(deref(lt.Yearly), rate)

	pd := utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(deref(lt.Daily))), rate)
	pw := utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(deref(lt.Weekly))), rate)
	pm := utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(lt.Monthly)), rate)
	py := utils.CalculateAdaptiveCost(float64(utils.CalculateProratedCost(deref(lt.Yearly))), rate)

	out := map[string]any{
		"id":         lt.Id,
		"created_at": lt.CreatedAt,
		"updated_at": lt.UpdatedAt,
		"key":        lt.Key,
		"name":       lt.Name,
		"desc":       lt.Desc,
		"detail":     lt.Detail,
		"type":       lt.Type,
		"currency":   lt.Currency,
		"daily":      d,
		"weekly":     w,
		"monthly":    m,
		"yearly":     y,
		"prorated": map[string]any{
			"daily":   pd,
			"weekly":  pw,
			"monthly": pm,
			"yearly":  py,
		},
		"meta":   mergeMeta(lt.Meta, rate),
		"status": lt.Status,
	}
	return out
}

// deref returns the *float64 value or 0 when nil.
func deref(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// parseRate pulls { "rate": <num> } from a statics List Meta blob.
func parseRate(raw json.RawMessage, fallback float64) float64 {
	if len(raw) == 0 {
		return fallback
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return fallback
	}
	if v, ok := m["rate"].(float64); ok && v > 0 {
		return v
	}
	return fallback
}

// mergeMeta returns the entity's existing meta (parsed) merged with
// { currencyMeta: { rate } }. Mirrors TS `meta: { ...entity.meta, currencyMeta: meta }`.
func mergeMeta(raw *string, rate float64) map[string]any {
	out := map[string]any{}
	if raw != nil && *raw != "" {
		_ = json.Unmarshal([]byte(*raw), &out)
	}
	out["currencyMeta"] = map[string]any{"rate": rate}
	return out
}

// sortByMonthly mirrors TS .sort((a, b) => a.monthly - b.monthly). We compare
// the int64 monthly values written by adjust().
func sortByMonthly(rows []map[string]any) {
	// Simple insertion sort — afterList row counts are small (license types
	// typically < 20) so the constant-factor overhead is negligible.
	for i := 1; i < len(rows); i++ {
		j := i
		for j > 0 && toInt64(rows[j-1]["monthly"]) > toInt64(rows[j]["monthly"]) {
			rows[j-1], rows[j] = rows[j], rows[j-1]
			j--
		}
	}
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	}
	return 0
}
