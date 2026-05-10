package formfield

import "github.com/thescaffold/gox-packages/libs/core/crud"

type FormFieldController struct {
	crud.CrudResource[FormField, CreateFormFieldDto, UpdateFormFieldDto]
	entity *FormFieldEntity `inject:""`
}

func (c *FormFieldController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FormField, CreateFormFieldDto, UpdateFormFieldDto]{
		Name:       "FormFormField",
		Searchable: []string{"key", "name"},
	})
}
