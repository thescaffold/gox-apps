package sinktype

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// SinkTypeController mirrors ntx-apps/libs/polylog/src/api/sink-type/sink-type.controller.ts.
type SinkTypeController struct {
	crud.CrudResource[SinkType, CreateSinkTypeDto, UpdateSinkTypeDto]

	entity *SinkTypeEntity `inject:""`
	lang   *i18n.Service   `inject:""`
}

func (c *SinkTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[SinkType, CreateSinkTypeDto, UpdateSinkTypeDto]{
		// Mirrors TS sink-type.controller.ts:18 name = 'sink group'.
		Name: "sink group",
		// Mirrors TS sink-type.controller.ts:19 searchable = ['name','desc','tags'].
		Searchable: []string{"name", "desc", "tags"},
		// Mirrors TS sink-type.controller.ts:21 unique = ({name}) => [{name}].
		Unique: func(d *CreateSinkTypeDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
