package invite

import (
	"errors"

	flagsapp "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// InviteController mirrors ntx-apps/libs/identity/src/api/invite/invite.controller.ts.
// Ported behaviour:
//   - morphs.beforeCreate: hydrate userId/clientId/workspaceId from context.
//   - hooks.beforeCreate: enforce per-workspace invite count limit (TS uses
//     FlagService.status; gox uses a static cap from defaultInviteLimit until
//     HTTP-wired).
//   - hooks.afterDelete: cascade soft-delete on the user-client-workspace row
//     bound to the invite's owner triple.
type InviteController struct {
	crud.CrudResource[Invite, CreateInviteDto, UpdateInviteDto]

	entity    *InviteEntity                          `inject:""`
	ucwEntity *identityucw.UserClientWorkspaceEntity `inject:""`
	flags     *flagsapp.FlagService                  `inject:""`
	lang      *i18n.Service                          `inject:""`
}

const defaultInviteLimit = 50

func (c *InviteController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Invite, CreateInviteDto, UpdateInviteDto]{
		Name:       "invite",
		Searchable: []string{"ref"},
		Unique: func(d *CreateInviteDto) []utils.KeyValue {
			return []utils.KeyValue{{
				"workspace_id": d.WorkspaceId,
				"email":        d.Email,
			}}
		},
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateInviteDto)
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
				return p, nil
			},
		},
		Hooks: map[string]crud.HookFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) error {
				p, ok := payload.(*CreateInviteDto)
				if !ok || p == nil || c.entity == nil {
					return nil
				}
				// TS-aligned: FlagService.status('apps.limits.user') gates first;
				// fall back to a static per-workspace cap so brand-new clients
				// without a registered flag row still see a sane limit.
				if c.flags != nil && !c.flags.Status("apps.limits.user", "user", p.UserId, p.ClientId, p.WorkspaceId) {
					return errors.New("apps.identity.invite.limit.error")
				}
				count, _ := c.entity.Count(`"workspace_id" = ?`, p.WorkspaceId)
				if count >= int64(defaultInviteLimit) {
					return errors.New("apps.identity.invite.limit.error")
				}
				return nil
			},
			crud.AfterDelete: func(payload any, ctx ntxctx.NTXContext) error {
				p, ok := payload.(*Invite)
				if !ok || p == nil || c.ucwEntity == nil {
					return nil
				}
				// Cascade: remove the user-client-workspace row that mirrors the
				// invite's (userId, clientId, workspaceId) triple. gox doesn't
				// distinguish member vs owner on the UCW row directly — the
				// soft-delete sweeps any matching row.
				_, _ = c.ucwEntity.Delete(
					`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ?`,
					p.UserId, p.ClientId, p.WorkspaceId,
				)
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}
