package flag

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FlagController mirrors ntx-apps/libs/flags/src/api/flag/flag.controller.ts.
type FlagController struct {
	crud.CrudResource[Flag, CreateFlagDto, UpdateFlagDto]
	entity *FlagEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *FlagController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Flag, CreateFlagDto, UpdateFlagDto]{
		// Mirrors TS flag.controller.ts:19 name = 'flag'.
		Name: "flag",
		// Mirrors TS flag.controller.ts:20 searchable = [].
		Searchable: nil,
		// TS unique = (entity: Flag) => [] — never blocks creation.
		// Mirrors TS morphs.beforeCreate: spreads { userId, clientId,
		// workspaceId } from the request context onto the payload.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateFlagDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok {
						p.WorkspaceId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
