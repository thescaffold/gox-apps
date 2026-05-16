package sourcetype

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// SourceTypeController mirrors ntx-apps/libs/polylog/src/api/source-type/source-type.controller.ts.
type SourceTypeController struct {
	crud.CrudResource[SourceType, CreateSourceTypeDto, UpdateSourceTypeDto]

	entity *SourceTypeEntity `inject:""`
	lang   *i18n.Service     `inject:""`
}

func (c *SourceTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[SourceType, CreateSourceTypeDto, UpdateSourceTypeDto]{
		// Mirrors TS source-type.controller.ts:18 name = 'source group'.
		Name: "source group",
		// Mirrors TS source-type.controller.ts:19 searchable = ['name','desc','tags'].
		Searchable: []string{"name", "desc", "tags"},
		// Mirrors TS source-type.controller.ts:21 unique = ({name}) => [{name}].
		Unique: func(d *CreateSourceTypeDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
