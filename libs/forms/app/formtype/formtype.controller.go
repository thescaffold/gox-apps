package formtype

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FormTypeController mirrors ntx-apps/libs/forms/src/api/form-type/form-type.controller.ts.
type FormTypeController struct {
	crud.CrudResource[FormType, CreateFormTypeDto, UpdateFormTypeDto]
	entity *FormTypeEntity `inject:""`
	lang   *i18n.Service   `inject:""`
}

func (c *FormTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FormType, CreateFormTypeDto, UpdateFormTypeDto]{
		// Mirrors TS form-type.controller.ts:18 name = 'form group'.
		Name: "form group",
		// Mirrors TS form-type.controller.ts:19 searchable = ['name','desc','tags'].
		Searchable: []string{"name", "desc", "tags"},
		// Mirrors TS form-type.controller.ts:21 unique = ({ name }) => [{ name }].
		Unique: func(d *CreateFormTypeDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
