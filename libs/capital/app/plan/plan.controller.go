package plan

import (
	"encoding/json"
	"errors"
	"fmt"

	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	flagsapp "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/services/marker"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// PlanController mirrors ntx-apps/libs/capital/src/api/plan/plan.controller.ts.
// Now implements:
//   - morphs.beforeCreate: hydrate ownership triple from context + default
//     planType to "lite" + periodType to "monthly"
//   - hooks.beforeCreate: validate planType exists, validate the user has a
//     matching paid marker (MarkerService verify) before allowing plan creation
//   - hooks.afterCreate: register the planType.meta.flags via FlagService
//     (wired in Phase E).
type PlanController struct {
	crud.CrudResource[Plan, CreatePlanDto, UpdatePlanDto]

	entity         *PlanEntity                       `inject:""`
	planTypeEntity *capitalplantype.PlanTypeEntity   `inject:""`
	markerService  *marker.Service                   `inject:""`
	flagService    *flagsapp.FlagService             `inject:""`
	lang           *i18n.Service                     `inject:""`
}

func (c *PlanController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Plan, CreatePlanDto, UpdatePlanDto]{
		Name:       "plan",
		Searchable: []string{},
		Unique: func(d *CreatePlanDto) []utils.KeyValue {
			return []utils.KeyValue{{
				"workspace_id": d.WorkspaceId,
				"type_id":      d.TypeId,
			}}
		},
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreatePlanDto)
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
				// Default typeId → planType "lite" when unspecified.
				if p.TypeId == "" && c.planTypeEntity != nil {
					if lite, _ := c.planTypeEntity.First(`"key" = ?`, "lite"); lite != nil {
						p.TypeId = lite.Id
					}
				}
				// Default periodType → monthly.
				if p.PeriodType == nil || *p.PeriodType == "" {
					pt := "monthly"
					p.PeriodType = &pt
				}
				return p, nil
			},
		},
		Hooks: map[string]crud.HookFn{
			// beforeCreate — validate the planType + matching paid marker.
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) error {
				p, ok := payload.(*CreatePlanDto)
				if !ok || p == nil {
					return nil
				}
				if c.planTypeEntity == nil {
					return nil
				}
				period := "monthly"
				if p.PeriodType != nil && *p.PeriodType != "" {
					period = *p.PeriodType
				}
				pt, _ := c.planTypeEntity.First(`"id" = ?`, p.TypeId)
				if pt == nil {
					return errors.New("apps.capital.app.post.plan.error.invalid-plan-type")
				}
				amount := planTypeAmount(pt, period)
				if amount == 0 {
					return errors.New("apps.capital.app.post.plan.error.invalid-plan-type")
				}
				// MarkerService.verify — matches TS markerService.verify
				// (`payment:<user>:<workspace>`, amount).
				if c.markerService != nil {
					userId, _, workspaceId := triple(ctx)
					key := fmt.Sprintf("payment:%s:%s", userId, workspaceId)
					if !c.markerService.Verify(key, fmt.Sprintf("%d", amount), true) {
						return errors.New("apps.capital.app.post.plan.error.payment-not-verified")
					}
				}
				return nil
			},
			// afterCreate — register planType.meta.flags via FlagService.
			// Mirrors TS hooks.afterCreate which seeds the user's per-flag rows
			// out of planType.meta.flags. Throws on failure (matches TS).
			crud.AfterCreate: func(payload any, ctx ntxctx.NTXContext) error {
				p, ok := payload.(*Plan)
				if !ok || p == nil || c.flagService == nil || c.planTypeEntity == nil {
					return nil
				}
				pt, _ := c.planTypeEntity.First(`"id" = ?`, p.TypeId)
				if pt == nil || len(pt.Meta) == 0 {
					return nil
				}
				var meta map[string]any
				if err := json.Unmarshal(pt.Meta, &meta); err != nil {
					return nil
				}
				flagsMap, _ := meta["flags"].(map[string]any)
				if len(flagsMap) == 0 {
					return nil
				}
				dtos := flagsapp.FromMeta(p.UserId, p.ClientId, p.WorkspaceId, flagsMap)
				ok2, err := c.flagService.Register(dtos)
				if err != nil {
					return err
				}
				if !ok2 {
					return errors.New("Failed to register flags for the user plan, try again.")
				}
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}

// planTypeAmount returns the period-specific monthly/yearly/etc. cost.
func planTypeAmount(pt *capitalplantype.PlanType, periodType string) int {
	if pt == nil {
		return 0
	}
	switch periodType {
	case "daily":
		if pt.Daily != nil {
			return *pt.Daily
		}
	case "weekly":
		if pt.Weekly != nil {
			return *pt.Weekly
		}
	case "monthly":
		return pt.Monthly
	case "yearly":
		if pt.Yearly != nil {
			return *pt.Yearly
		}
	}
	return 0
}

// triple extracts (userId, clientId, workspaceId) from NTXContext.
func triple(ntx ntxctx.NTXContext) (string, string, string) {
	userId, clientId, workspaceId := "", "", ""
	if ntx.User != nil {
		userId, _ = ntx.User["id"].(string)
	}
	if userId == "" {
		userId = ntx.UserID
	}
	if ntx.Client != nil {
		clientId, _ = ntx.Client["id"].(string)
	}
	if clientId == "" {
		clientId = ntx.ClientID
	}
	if ntx.Workspace != nil {
		workspaceId, _ = ntx.Workspace["id"].(string)
	}
	if workspaceId == "" {
		workspaceId = ntx.WorkspaceID
	}
	return userId, clientId, workspaceId
}
