package payment

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// PaymentController mirrors ntx-apps/libs/capital/src/api/payment/payment.controller.ts.
type PaymentController struct {
	crud.CrudResource[Payment, CreatePaymentDto, UpdatePaymentDto]

	entity *PaymentEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *PaymentController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Payment, CreatePaymentDto, UpdatePaymentDto]{
		// Mirrors TS payment.controller.ts name = 'payment'.
		Name: "payment",
		// Mirrors TS payment.controller.ts searchable = ['reference'].
		Searchable: []string{"reference"},
		// Mirrors TS payment.controller.ts unique = ({reference}) => [{reference}].
		Unique: func(d *CreatePaymentDto) []map[string]any {
			return []map[string]any{{"reference": d.Reference}}
		},
		// Mirrors TS payment.controller.ts morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreatePaymentDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
