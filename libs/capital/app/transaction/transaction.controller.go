package transaction

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// TransactionController mirrors ntx-apps/libs/capital/src/api/transaction/transaction.controller.ts.
type TransactionController struct {
	crud.CrudResource[Transaction, CreateTransactionDto, UpdateTransactionDto]

	entity *TransactionEntity `inject:""`
	lang   *i18n.Service      `inject:""`
}

func (c *TransactionController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Transaction, CreateTransactionDto, UpdateTransactionDto]{
		// Mirrors TS transaction.controller.ts name = 'transaction'.
		Name: "transaction",
		// Mirrors TS transaction.controller.ts searchable = ['narration','desc'].
		Searchable: []string{"narration", "desc"},
		// Mirrors TS transaction.controller.ts unique = ({reference}) =>
		// [{reference}].
		Unique: func(d *CreateTransactionDto) []map[string]any {
			return []map[string]any{{"reference": d.Reference}}
		},
		// Mirrors TS transaction.controller.ts morphs.beforeCreate. TS only
		// spreads userId — clientId/workspaceId aren't on the Transaction
		// entity — so we mirror that exactly.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateTransactionDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
