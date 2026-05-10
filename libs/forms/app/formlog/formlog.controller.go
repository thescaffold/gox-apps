package formlog

import "github.com/thescaffold/gox-packages/libs/core/crud"

type FormLogController struct {
	crud.CrudResource[FormLog, CreateFormLogDto, UpdateFormLogDto]
	entity *FormLogEntity `inject:""`
}

func (c *FormLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FormLog, CreateFormLogDto, UpdateFormLogDto]{
		Name:       "FormFormLog",
		Searchable: []string{"form_id", "form_field_id", "key"},
	})
}
