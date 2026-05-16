package token

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// TokenController mirrors ntx-apps/libs/fuss/src/api/token/token.controller.ts.
type TokenController struct {
	crud.CrudResource[Token, CreateTokenDto, UpdateTokenDto]

	entity *TokenEntity  `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *TokenController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Token, CreateTokenDto, UpdateTokenDto]{
		// Mirrors TS token.controller.ts:19 name = 'token'.
		Name: "token",
		// Mirrors TS token.controller.ts:20 searchable = ['service'].
		Searchable: []string{"service"},
		// TS unique = (entity: Token) => [] — never blocks creation.
		// Mirrors TS morphs.beforeCreate: spreads { userId, clientId,
		// workspaceId } from the request context onto the payload.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateTokenDto)
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
