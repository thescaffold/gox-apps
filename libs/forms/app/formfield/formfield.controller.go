package formfield

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FormFieldController mirrors ntx-apps/libs/forms/src/api/form-field/form-field.controller.ts.
type FormFieldController struct {
	crud.CrudResource[FormField, CreateFormFieldDto, UpdateFormFieldDto]
	entity *FormFieldEntity `inject:""`
	lang   *i18n.Service    `inject:""`
}

func (c *FormFieldController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FormField, CreateFormFieldDto, UpdateFormFieldDto]{
		// Mirrors TS form-field.controller.ts:18 name = 'form field'.
		Name: "form field",
		// Mirrors TS form-field.controller.ts:19 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS form-field.controller.ts:21-24:
		//   unique = ({ formId, key, name }) => [{ formId, key }, { formId, name }]
		Unique: func(d *CreateFormFieldDto) []map[string]any {
			return []map[string]any{
				{"form_id": d.FormId, "key": d.Key},
				{"form_id": d.FormId, "name": d.Name},
			}
		},
	})
	c.SetLang(c.lang)
}
