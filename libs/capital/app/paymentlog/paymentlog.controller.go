package paymentlog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// PaymentLogController mirrors ntx-apps/libs/capital/src/api/payment-log/payment-log.controller.ts.
type PaymentLogController struct {
	crud.CrudResource[PaymentLog, CreatePaymentLogDto, UpdatePaymentLogDto]

	entity *PaymentLogEntity `inject:""`
	lang   *i18n.Service     `inject:""`
}

func (c *PaymentLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[PaymentLog, CreatePaymentLogDto, UpdatePaymentLogDto]{
		// Mirrors TS payment-log.controller.ts name = 'payment log'.
		Name: "payment log",
		// Mirrors TS payment-log.controller.ts searchable = [] (empty).
		Searchable: []string{},
		// TS unique = () => [] — no uniqueness constraints; left unset.
	})
	c.SetLang(c.lang)
}
