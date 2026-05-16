package account

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// AccountController mirrors ntx-apps/libs/capital/src/api/account/account.controller.ts.
type AccountController struct {
	crud.CrudResource[Account, CreateAccountDto, UpdateAccountDto]

	entity *AccountEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *AccountController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Account, CreateAccountDto, UpdateAccountDto]{
		// Mirrors TS account.controller.ts:21 name = 'account'.
		Name: "account",
		// Mirrors TS account.controller.ts:22 searchable = ['desc','label'].
		Searchable: []string{"desc", "label"},
		// Mirrors TS account.controller.ts:24-27 unique = ({reference,number}) =>
		// [{reference},{number}] — two OR'd constraints.
		Unique: func(d *CreateAccountDto) []map[string]any {
			out := []map[string]any{{"reference": d.Reference}}
			if d.Number != nil {
				out = append(out, map[string]any{"number": *d.Number})
			}
			return out
		},
		// Mirrors TS account.controller.ts morphs.beforeCreate that spreads
		// userId/clientId/workspaceId from context.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateAccountDto)
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
	})
	c.SetLang(c.lang)
}
