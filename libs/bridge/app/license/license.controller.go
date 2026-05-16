package license

import (
	"encoding/json"

	"github.com/thescaffold/gox-apps/libs/bridge/app/plantype"
	"github.com/thescaffold/gox-apps/libs/bridge/app/preference"
	"github.com/thescaffold/gox-apps/libs/bridge/app/webhook"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// LicensePreferenceType mirrors TS LicensePreferenceType in app.config.ts.
const (
	LicensePreferenceTypeEnv  = "env"
	LicensePreferenceTypeFlag = "flag"
)

// LicenseStatusPending mirrors TS LicenseStatusType.Pending.
const LicenseStatusPending = "pending"

// LicenseController mirrors ntx-apps/libs/bridge/src/api/license/license.controller.ts.
// Phase 7.2 wires the BeforeCreate/BeforeUpdate morphs that flatten the
// nested {type, preferences, planTypes, webhooks, amount} body onto the
// License entity row (with the nested data serialised into Meta), and the
// AfterCreate/AfterUpdate/AfterDelete hooks that drive child-row sync via
// PreferenceService/PlanTypeService/WebhookService. AfterCreate also calls
// capital.UsageService.StartUsage (and AfterDelete calls StopUsage), so the
// licence's usage row tracks the lifecycle for downstream billing.
type LicenseController struct {
	crud.CrudResource[License, CreateLicenseDto, UpdateLicenseDto]

	entity            *LicenseEntity                `inject:""`
	preferenceService *preference.PreferenceService `inject:""`
	planTypeService   *plantype.PlanTypeService     `inject:""`
	webhookService    *webhook.WebhookService       `inject:""`
	usageService      *capitalusage.UsageService    `inject:""`
	lang              *i18n.Service                 `inject:""`
}

func (c *LicenseController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[License, CreateLicenseDto, UpdateLicenseDto]{
		// Mirrors TS license.controller.ts:28 name = 'license'.
		Name: "license",
		// Mirrors TS license.controller.ts:29 searchable = [] (empty).
		Searchable: []string{},
		// Mirrors TS license.controller.ts:33-58 morphs.beforeCreate: flatten
		// {userId, clientId, workspaceId, typeId, periodType, meta, status}
		// where meta = JSON({type, preferences, planTypes, webhooks, amount}).
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateLicenseDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = id
					}
				}
				p.TypeId = p.Type.Id
				if p.PeriodType == nil || *p.PeriodType == "" {
					monthly := "monthly"
					p.PeriodType = &monthly
				}
				nested := map[string]any{
					"type":        p.Type,
					"preferences": p.Preferences,
					"planTypes":   p.PlanTypes,
					"webhooks":    p.Webhooks,
					"amount":      p.Amount,
				}
				if b, err := json.Marshal(nested); err == nil {
					s := string(b)
					p.Meta = &s
				}
				status := LicenseStatusPending
				p.Status = &status
				return p, nil
			},
			// Mirrors TS license.controller.ts:59-76 morphs.beforeUpdate: same
			// flattening for {type, preferences, planTypes, webhooks, amount}
			// but no user/client/workspace re-injection (the row's owner is
			// already set).
			crud.BeforeUpdate: func(payload any, _ ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*UpdateLicenseDto)
				if !ok || p == nil {
					return nil, nil
				}
				nested := map[string]any{
					"type":        p.Type,
					"preferences": p.Preferences,
					"planTypes":   p.PlanTypes,
					"webhooks":    p.Webhooks,
					"amount":      p.Amount,
				}
				if b, err := json.Marshal(nested); err == nil {
					s := string(b)
					p.Meta = &s
				}
				return p, nil
			},
		},
		// Mirrors TS license.controller.ts:78-155 hooks.afterCreate/afterUpdate/
		// afterDelete: drive child-row sync via process() + fire the capital
		// UsageService start/stop events so billing tracks the lifecycle.
		Hooks: map[string]crud.HookFn{
			crud.AfterCreate: func(payload any, _ ntxctx.NTXContext) error {
				c.process(payload)
				c.fireUsage(payload, true)
				return nil
			},
			crud.AfterUpdate: func(payload any, _ ntxctx.NTXContext) error {
				c.process(payload)
				return nil
			},
			crud.AfterDelete: func(payload any, _ ntxctx.NTXContext) error {
				c.process(payload)
				c.fireUsage(payload, false)
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}

// process syncs the License's nested {preferences, planTypes, webhooks} onto
// their respective child tables. Mirrors TS license.controller.ts:171-258.
//
// Inputs may be either *License (the entity instance handed back by the crud
// AfterCreate/AfterUpdate/AfterDelete morphs) or anything else — non-License
// payloads are silently ignored. The Meta column is JSON-decoded back to the
// nested shape so we know which preferences/plan-types/webhooks to write.
// fireUsage emits the capital usage start/stop event for the License's
// (user, client, workspace) tuple. Mirrors TS UsageService.startUsage on
// AfterCreate and stopUsage on AfterDelete — keyed by License id so capital
// can attribute usage back to the correct licence row.
func (c *LicenseController) fireUsage(payload any, start bool) {
	if c.usageService == nil {
		return
	}
	lic, ok := payload.(*License)
	if !ok || lic == nil {
		return
	}
	name := "apps/bridge/license/" + lic.Id
	meta := map[string]any{"licenseId": lic.Id}
	if start {
		c.usageService.StartUsage(lic.UserId, lic.ClientId, lic.WorkspaceId, name, lic.Id, meta, 1)
		return
	}
	c.usageService.StopUsage(lic.UserId, lic.ClientId, lic.WorkspaceId, name, lic.Id, meta)
}

func (c *LicenseController) process(payload any) {
	lic, ok := payload.(*License)
	if !ok || lic == nil || lic.Meta == nil || *lic.Meta == "" {
		return
	}
	var nested struct {
		Type        any             `json:"type"`
		Preferences []PreferenceRef `json:"preferences"`
		PlanTypes   []PlanTypeRef   `json:"planTypes"`
		Webhooks    []WebhookRef    `json:"webhooks"`
		Amount      *float64        `json:"amount"`
	}
	if err := json.Unmarshal([]byte(*lic.Meta), &nested); err != nil {
		return
	}

	// Sync env preferences: keep prior values, layer new overrides on top,
	// then rewrite the env subset.
	if c.preferenceService != nil {
		envOverrides := map[string]string{}
		flagOverrides := map[string]string{}
		for _, p := range nested.Preferences {
			val := ""
			if p.Value != nil {
				val = *p.Value
			}
			switch p.Type {
			case LicensePreferenceTypeEnv:
				envOverrides[p.Key] = val
			case LicensePreferenceTypeFlag:
				flagOverrides[p.Key] = val
			}
		}
		// env: read existing → merge → wipe → reinsert.
		existing, _ := c.preferenceService.Find(map[string]any{
			"license_id": lic.Id,
			"type":       LicensePreferenceTypeEnv,
		})
		merged := map[string]string{}
		for _, e := range existing {
			if e.Value != nil {
				merged[e.Key] = *e.Value
			} else {
				merged[e.Key] = ""
			}
		}
		for k, v := range envOverrides {
			merged[k] = v
		}
		_, _ = c.preferenceService.Delete(map[string]any{
			"license_id": lic.Id,
			"type":       LicensePreferenceTypeEnv,
		})
		for k, v := range merged {
			val := v
			t := LicensePreferenceTypeEnv
			_ = c.preferenceService.Save(&preference.Preference{
				LicenseId: lic.Id,
				Key:       k,
				Value:     &val,
				Type:      &t,
			})
		}
		// flag: wipe + reinsert (TS doesn't preserve existing flags).
		_, _ = c.preferenceService.Delete(map[string]any{
			"license_id": lic.Id,
			"type":       LicensePreferenceTypeFlag,
		})
		for k, v := range flagOverrides {
			val := v
			t := LicensePreferenceTypeFlag
			_ = c.preferenceService.Save(&preference.Preference{
				LicenseId: lic.Id,
				Key:       k,
				Value:     &val,
				Type:      &t,
			})
		}
	}

	// Sync plan types: wipe + reinsert per TS.
	if c.planTypeService != nil {
		_, _ = c.planTypeService.Delete(map[string]any{"license_id": lic.Id})
		for _, pt := range nested.PlanTypes {
			key := slugify(pt.Name)
			row := &plantype.PlanType{
				LicenseId: lic.Id,
				Key:       key,
				Name:      pt.Name,
				Desc:      pt.Desc,
				Detail:    pt.Detail,
				Type:      pt.Type,
				Monthly:   pt.Monthly,
				Daily:     pt.Daily,
				Weekly:    pt.Weekly,
				Yearly:    pt.Yearly,
			}
			if pt.Currency != nil {
				row.Currency = *pt.Currency
			}
			if len(pt.Flags) > 0 {
				if mb, err := json.Marshal(map[string]any{"flags": pt.Flags}); err == nil {
					row.Meta = mb
				}
			}
			_ = c.planTypeService.Save(row)
		}
	}

	// Sync webhooks: wipe + reinsert per TS.
	if c.webhookService != nil {
		_, _ = c.webhookService.Delete(map[string]any{"license_id": lic.Id})
		for _, wh := range nested.Webhooks {
			events := map[string]bool{}
			for _, e := range wh.Events {
				events[e] = true
			}
			meta, _ := json.Marshal(map[string]any{"events": events})
			typ := "POST"
			url := wh.Url
			_ = c.webhookService.Save(&webhook.Webhook{
				LicenseId: lic.Id,
				Type:      &typ,
				Url:       &url,
				Meta:      meta,
			})
		}
	}
}

// slugify mirrors TS `name.toLowerCase().replace(/\s+/g, '-')` used as the
// PlanType key when none is supplied.
func slugify(s string) string {
	out := make([]byte, 0, len(s))
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			out = append(out, byte(r+32))
			prevDash = false
		case r == ' ' || r == '\t' || r == '\n':
			if !prevDash {
				out = append(out, '-')
				prevDash = true
			}
		default:
			out = append(out, byte(r))
			prevDash = false
		}
	}
	return string(out)
}
