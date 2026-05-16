package provider

import (
	"errors"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ProviderController mirrors ntx-apps/libs/capital/src/api/provider/provider.controller.ts.
// Implements the full 7-hook morph surface from TS:
//
//   - afterGet  : strip signature + token before responding
//   - afterList : strip signature + token from each row
//   - afterCreate: tracker.Message activity-update (cards +1)
//   - beforeUpdate: guard primary-toggle (must change; must have ≥2 cards)
//   - afterUpdate: when primary=true, demote sibling cards
//   - beforeDelete: guard delete (no primary; must have ≥2 cards)
//   - afterDelete: tracker.Message activity-update (cards -1)
type ProviderController struct {
	crud.CrudResource[Provider, CreateProviderDto, UpdateProviderDto]

	entity  *ProviderEntity        `inject:""`
	lang    *i18n.Service          `inject:""`
	tracker *events.TrackerService `inject:""`
}

func (c *ProviderController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Provider, CreateProviderDto, UpdateProviderDto]{
		Name:       "provider",
		Searchable: []string{"name", "email"},
		Morphs: map[string]crud.MorphFn{
			// afterGet — strip secrets from a single row.
			crud.AfterGet: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				if p, ok := payload.(*Provider); ok && p != nil {
					p.Signature = nil
					p.Token = nil
				}
				return payload, nil
			},
			// afterList — strip secrets from every row in the page.
			crud.AfterList: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				if rows, ok := payload.([]Provider); ok {
					for i := range rows {
						rows[i].Signature = nil
						rows[i].Token = nil
					}
					return rows, nil
				}
				if rows, ok := payload.([]*Provider); ok {
					for _, r := range rows {
						if r != nil {
							r.Signature = nil
							r.Token = nil
						}
					}
					return rows, nil
				}
				return payload, nil
			},
		},
		Hooks: map[string]crud.HookFn{
			// afterCreate — bump the user's `cards` activity counter.
			crud.AfterCreate: func(payload any, ctx ntxctx.NTXContext) error {
				if c.tracker == nil {
					return nil
				}
				userId, clientId, workspaceId := triple(ctx)
				c.tracker.Message("apps.identity.attribute.update", map[string]any{
					"type":        "activity",
					"key":         "cards",
					"value":       "1",
					"action":      "add",
					"userId":      userId,
					"clientId":    clientId,
					"workspaceId": workspaceId,
				})
				return nil
			},
			// beforeUpdate — when toggling `primary`, require the user to have
			// at least one fallback card so a chargeable provider always
			// remains. The TS "already-updated" guard (`existing.primary ===
			// payload.primary`) lives in the HooksCtx variant below since the
			// legacy HookFn doesn't receive the resolved entity.
			crud.BeforeUpdate: func(payload any, ctx ntxctx.NTXContext) error {
				dto, ok := payload.(*UpdateProviderDto)
				if !ok || dto == nil || dto.Primary == nil {
					return nil
				}
				userId, _, workspaceId := triple(ctx)
				count, _ := c.entity.Count(
					`"account_id" IN (SELECT "id" FROM "CapitalAccounts" WHERE "user_id" = ? AND "workspace_id" = ?)`,
					userId, workspaceId,
				)
				if count < 2 {
					return errors.New("apps.capital.provider.before-update.error.only-one-card")
				}
				return nil
			},
			// afterUpdate — promoting one to primary demotes all siblings.
			crud.AfterUpdate: func(payload any, ctx ntxctx.NTXContext) error {
				p, ok := payload.(*Provider)
				if !ok || p == nil || p.Primary == nil || !*p.Primary {
					return nil
				}
				userId, _, workspaceId := triple(ctx)
				notPrimary := false
				_, _ = c.entity.Update(
					&Provider{Primary: &notPrimary},
					`"id" <> ? AND "account_id" IN (SELECT "id" FROM "CapitalAccounts" WHERE "user_id" = ? AND "workspace_id" = ?)`,
					p.Id, userId, workspaceId,
				)
				return nil
			},
			// beforeDelete — can't delete primary; must have a fallback.
			crud.BeforeDelete: func(payload any, ctx ntxctx.NTXContext) error {
				idDto, ok := payload.(map[string]any)
				id := ""
				if ok {
					id, _ = idDto["id"].(string)
				}
				if id == "" {
					if p, ok := payload.(*Provider); ok && p != nil {
						id = p.Id
					}
				}
				if id == "" {
					return nil
				}
				existing, _ := c.entity.First(`"id" = ?`, id)
				if existing == nil {
					return nil
				}
				if existing.Primary != nil && *existing.Primary {
					return errors.New("apps.capital.provider.before-update.error.cannot-delete-primary")
				}
				userId, _, workspaceId := triple(ctx)
				count, _ := c.entity.Count(
					`"account_id" IN (SELECT "id" FROM "CapitalAccounts" WHERE "user_id" = ? AND "workspace_id" = ?)`,
					userId, workspaceId,
				)
				if count < 2 {
					return errors.New("apps.capital.provider.before-update.error.only-one-card")
				}
				return nil
			},
			// afterDelete — decrement the user's `cards` activity counter.
			crud.AfterDelete: func(payload any, ctx ntxctx.NTXContext) error {
				if c.tracker == nil {
					return nil
				}
				userId, clientId, workspaceId := triple(ctx)
				c.tracker.Message("apps.identity.attribute.update", map[string]any{
					"type":        "activity",
					"key":         "cards",
					"value":       "1",
					"action":      "remove",
					"userId":      userId,
					"clientId":    clientId,
					"workspaceId": workspaceId,
				})
				return nil
			},
		},
		HooksCtx: map[string]crud.HookCtxFn{
			// beforeUpdate (rich) — TS-equivalent "already-updated" guard.
			// Fires after the legacy BeforeUpdate hook with the resolved
			// entity, so we can compare prior vs. requested values.
			crud.BeforeUpdate: func(event crud.HookEvent, ctx ntxctx.NTXContext) error {
				dto, ok := event.DTO.(*UpdateProviderDto)
				if !ok || dto == nil || dto.Primary == nil {
					return nil
				}
				existing, ok := event.Entity.(*Provider)
				if !ok || existing == nil {
					return nil
				}
				if existing.Primary != nil && *existing.Primary == *dto.Primary {
					return errors.New("apps.capital.provider.before-update.error.already-updated")
				}
				return nil
			},
			// beforeDelete (rich) — TS-equivalent "can't delete primary" guard.
			// Uses the resolved entity to inspect the current primary flag.
			crud.BeforeDelete: func(event crud.HookEvent, ctx ntxctx.NTXContext) error {
				existing, ok := event.Entity.(*Provider)
				if !ok || existing == nil {
					return nil
				}
				if existing.Primary != nil && *existing.Primary {
					return errors.New("apps.capital.provider.before-update.error.cannot-delete-primary")
				}
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}

// triple extracts the (userId, clientId, workspaceId) identity tuple from
// NTXContext. Mirrors the pattern used across gox controllers.
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
