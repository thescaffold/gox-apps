package account

import "github.com/thescaffold/gox-packages/libs/core/crud"

type AccountController struct {
	crud.CrudResource[Account, CreateAccountDto, UpdateAccountDto]
	entity *AccountEntity `inject:""`
}

func (c *AccountController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Account, CreateAccountDto, UpdateAccountDto]{
		Name:       "CapitalAccount",
		Searchable: []string{"user_id", "client_id", "workspace_id", "reference"},
	})
}
