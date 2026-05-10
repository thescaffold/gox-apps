package formtype

import "github.com/thescaffold/gox-packages/libs/core/crud"

type FormTypeController struct {
	crud.CrudResource[FormType, CreateFormTypeDto, UpdateFormTypeDto]
	entity *FormTypeEntity `inject:""`
}

func (c *FormTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FormType, CreateFormTypeDto, UpdateFormTypeDto]{
		Name:       "FormFormType",
		Searchable: []string{"name"},
	})
}
