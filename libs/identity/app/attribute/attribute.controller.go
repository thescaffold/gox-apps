package attribute

import "github.com/thescaffold/gox-packages/libs/core/crud"

type AttributeController struct {
	crud.CrudResource[Attribute, CreateAttributeDto, UpdateAttributeDto]
	entity *AttributeEntity `inject:""`
}

func (c *AttributeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Attribute, CreateAttributeDto, UpdateAttributeDto]{
		Name:       "IdentityAttribute",
		Searchable: []string{"user_id", "client_id", "workspace_id", "key"},
	})
}
