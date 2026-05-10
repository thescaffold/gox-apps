package transaction

import "github.com/thescaffold/gox-packages/libs/core/crud"

type TransactionController struct {
	crud.CrudResource[Transaction, CreateTransactionDto, UpdateTransactionDto]
	entity *TransactionEntity `inject:""`
}

func (c *TransactionController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Transaction, CreateTransactionDto, UpdateTransactionDto]{
		Name:       "CapitalTransaction",
		Searchable: []string{"user_id", "account_id", "currency", "reference"},
	})
}
