package tokenlog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// TokenLogController mirrors ntx-apps/libs/fuss/src/api/token-log/token-log.controller.ts.
type TokenLogController struct {
	crud.CrudResource[TokenLog, CreateTokenLogDto, UpdateTokenLogDto]

	entity *TokenLogEntity `inject:""`
	lang   *i18n.Service   `inject:""`
}

func (c *TokenLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[TokenLog, CreateTokenLogDto, UpdateTokenLogDto]{
		// Mirrors TS token-log.controller.ts:18 name = 'token log'.
		Name: "token log",
		// Mirrors TS token-log.controller.ts:19 searchable = ['value'].
		Searchable: []string{"value"},
		// TS unique = (entity: TokenLog) => [] — never blocks creation.
	})
	c.SetLang(c.lang)
}
