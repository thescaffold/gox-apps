package tokenlog

import "github.com/thescaffold/gox-packages-core/crud"

type TokenLogController struct {
	crud.CrudResource[TokenLog, CreateTokenLogDto, UpdateTokenLogDto]

	entity *TokenLogEntity `inject:""`
}

func (c *TokenLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[TokenLog, CreateTokenLogDto, UpdateTokenLogDto]{
		Name:       "FussTokenLog",
		Searchable: []string{"token_id", "value"},
	})
}
