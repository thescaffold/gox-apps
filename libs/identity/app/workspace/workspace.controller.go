package workspace

import (
	"errors"

	assetsapp "github.com/thescaffold/gox-apps/libs/assets/app"
	flagsapp "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// WorkspaceController mirrors ntx-apps/libs/identity/src/api/workspace/workspace.controller.ts.
// Ported behavior:
//   - morphs.beforeCreate: hydrate userId/clientId from context; supply default
//     logoUrl when missing (TS uses thumbnailUrl + assetsAppService.getDynamicAsset;
//     gox uses a placeholder URL until cross-app injection lands).
//   - hooks.beforeCreate: enforce a per-user workspace count limit (TS gates
//     via FlagService.status; gox uses a static limit until HTTP-wired).
//   - hooks.afterCreate: persist UserClientWorkspace owner row + emit
//     apps.flags.flag.log via TrackerService.
//   - hooks.beforeDelete: require ≥2 workspaces; refuse delete on the active
//     workspace.
type WorkspaceController struct {
	crud.CrudResource[Workspace, CreateWorkspaceDto, UpdateWorkspaceDto]

	entity    *WorkspaceEntity                       `inject:""`
	ucwEntity *identityucw.UserClientWorkspaceEntity `inject:""`
	tracker   *events.TrackerService                 `inject:""`
	assets    *assetsapp.AppService                  `inject:""`
	flags     *flagsapp.FlagService                  `inject:""`
	lang      *i18n.Service                          `inject:""`
}

const defaultWorkspaceLimit = 10

func (c *WorkspaceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Workspace, CreateWorkspaceDto, UpdateWorkspaceDto]{
		Name:       "workspace",
		Searchable: []string{"name", "desc"},
		Unique: func(d *CreateWorkspaceDto) []utils.KeyValue {
			return []utils.KeyValue{{
				"name":      d.Name,
				"user_id":   d.UserId,
				"client_id": d.ClientId,
			}}
		},
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateWorkspaceDto)
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
				if p.LogoUrl == nil || *p.LogoUrl == "" {
					// Mirrors TS assetsAppService.getDynamicAsset() — generates
					// a per-workspace placeholder thumbnail when none supplied.
					if c.assets != nil {
						if asset := c.assets.GetDynamicAsset(); asset != nil && asset.Url != "" {
							url := asset.Url
							p.LogoUrl = &url
						}
					}
					if p.LogoUrl == nil || *p.LogoUrl == "" {
						ph := "/assets/workspace-default.svg"
						p.LogoUrl = &ph
					}
				}
				return p, nil
			},
		},
		Hooks: map[string]crud.HookFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) error {
				userId, clientId, workspaceId := workspaceCtxTriple(ctx)
				if userId == "" || c.ucwEntity == nil {
					return nil
				}
				// TS-aligned: gate first via FlagService.status('apps.limits.user').
				// When the flag is inactive (cap reached), refuse the create.
				if c.flags != nil && !c.flags.Status("apps.limits.user", "user", userId, clientId, workspaceId) {
					return errors.New("apps.identity.workspace.limit.error")
				}
				// Fallback static cap so brand-new clients without registered
				// flag rows still get a sane upper bound.
				count, _ := c.ucwEntity.Count(`"user_id" = ? AND "client_id" = ?`, userId, clientId)
				if count >= int64(defaultWorkspaceLimit) {
					return errors.New("apps.identity.workspace.limit.error")
				}
				return nil
			},
			crud.AfterCreate: func(payload any, ctx ntxctx.NTXContext) error {
				w, ok := payload.(*Workspace)
				if !ok || w == nil || c.ucwEntity == nil {
					return nil
				}
				ownerType := "owner"
				ucw := &identityucw.UserClientWorkspace{
					UserId:      w.UserId,
					ClientId:    w.ClientId,
					WorkspaceId: w.Id,
					Status:      &ownerType,
				}
				_ = c.ucwEntity.Insert(ucw)
				if c.tracker != nil {
					c.tracker.Message("apps.flags.flag.log", map[string]any{
						"flag": map[string]any{
							"userId":      w.UserId,
							"clientId":    w.ClientId,
							"workspaceId": w.Id,
							"name":        "apps.limits.workspace",
							"level":       "user",
						},
						"limit": 1,
					})
				}
				return nil
			},
			crud.BeforeDelete: func(payload any, ctx ntxctx.NTXContext) error {
				if c.ucwEntity == nil {
					return nil
				}
				userId, _, _ := workspaceCtxTriple(ctx)
				count, _ := c.ucwEntity.Count(`"user_id" = ?`, userId)
				if count < 2 {
					return errors.New("apps.identity.workspace.hooks.after-create.minimum")
				}
				return nil
			},
		},
		HooksCtx: map[string]crud.HookCtxFn{
			// beforeDelete (rich) — TS-equivalent active-workspace guard
			// using the URL :id from HookEvent.
			crud.BeforeDelete: func(event crud.HookEvent, ctx ntxctx.NTXContext) error {
				_, _, activeWorkspaceID := workspaceCtxTriple(ctx)
				if event.ID != "" && event.ID == activeWorkspaceID {
					return errors.New("apps.identity.workspace.hooks.after-create.active")
				}
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}

func workspaceCtxTriple(ntx ntxctx.NTXContext) (string, string, string) {
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
