package template

import "github.com/thescaffold/gox-packages-core/crud"

type TemplateController struct {
	crud.CrudResource[Template, CreateTemplateDto, UpdateTemplateDto]
	entity *TemplateEntity `inject:""`
}

func (c *TemplateController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Template, CreateTemplateDto, UpdateTemplateDto]{
		Name:       "NotificationTemplate",
		Searchable: []string{"group", "key"},
	})
}
