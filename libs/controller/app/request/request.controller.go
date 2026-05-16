package request

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// RequestController mirrors ntx-apps/libs/controller/src/api/request/request.controller.ts.
// Despite the entity table being "ControllerRequests", the TS controller's
// human-facing `name` is "source log" (used in translated response titles),
// and `searchable` is empty.
type RequestController struct {
	crud.CrudResource[Request, CreateRequestDto, UpdateRequestDto]

	entity *RequestEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *RequestController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Request, CreateRequestDto, UpdateRequestDto]{
		// Mirrors TS request.controller.ts:18 name = 'source log'.
		Name: "source log",
		// Mirrors TS request.controller.ts:19 searchable = [].
		Searchable: nil,
	})
	c.SetLang(c.lang)
}
