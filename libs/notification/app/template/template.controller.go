package template

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// TemplateController mirrors ntx-apps/libs/notification/src/api/template/template.controller.ts.
type TemplateController struct {
	crud.CrudResource[Template, CreateTemplateDto, UpdateTemplateDto]

	entity *TemplateEntity `inject:""`
	lang   *i18n.Service   `inject:""`
}

func (c *TemplateController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Template, CreateTemplateDto, UpdateTemplateDto]{
		// Mirrors TS template.controller.ts:18 name = 'template'.
		Name: "template",
		// Mirrors TS template.controller.ts:19 searchable = ['group','key'].
		Searchable: []string{"group", "key"},
		// Mirrors TS template.controller.ts:21 unique = ({group,key}) => [{group,key}].
		Unique: func(d *CreateTemplateDto) []map[string]any {
			return []map[string]any{{"group": d.Group, "key": d.Key}}
		},
	})
	c.SetLang(c.lang)
}
