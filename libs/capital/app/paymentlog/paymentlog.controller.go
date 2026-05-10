package paymentlog

import "github.com/thescaffold/gox-packages/libs/core/crud"

type PaymentLogController struct {
	crud.CrudResource[PaymentLog, CreatePaymentLogDto, UpdatePaymentLogDto]
	entity *PaymentLogEntity `inject:""`
}

func (c *PaymentLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[PaymentLog, CreatePaymentLogDto, UpdatePaymentLogDto]{
		Name: "CapitalPaymentLog", Searchable: []string{"payment_id", "type"},
	})
}
