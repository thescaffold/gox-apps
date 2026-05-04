package form

import "github.com/thescaffold/gox-packages-core/crud"

type FormController struct {
	crud.CrudResource[Form, CreateFormDto, UpdateFormDto]
	entity *FormEntity `inject:""`
}

func (c *FormController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Form, CreateFormDto, UpdateFormDto]{
		Name:       "FormForm",
		Searchable: []string{"key", "name"},
	})
}
