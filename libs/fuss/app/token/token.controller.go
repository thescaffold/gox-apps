package token

import "github.com/thescaffold/gox-packages-core/crud"

type TokenController struct {
	crud.CrudResource[Token, CreateTokenDto, UpdateTokenDto]

	entity *TokenEntity `inject:""`
}

func (c *TokenController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Token, CreateTokenDto, UpdateTokenDto]{
		Name: "token",
		// Mirrors TS token.controller.ts:20 searchable = ['service'].
		Searchable: []string{"service"},
	})
}
