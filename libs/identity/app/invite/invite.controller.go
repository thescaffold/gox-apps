package invite

import "github.com/thescaffold/gox-packages-core/crud"

type InviteController struct {
	crud.CrudResource[Invite, CreateInviteDto, UpdateInviteDto]
	entity *InviteEntity `inject:""`
}

func (c *InviteController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Invite, CreateInviteDto, UpdateInviteDto]{
		Name:       "IdentityInvite",
		Searchable: []string{"user_id", "client_id", "workspace_id", "email"},
	})
}
