package payment
import "github.com/thescaffold/gox-packages-core/crud"
type PaymentController struct {
	crud.CrudResource[Payment, CreatePaymentDto, UpdatePaymentDto]
	entity *PaymentEntity `inject:""`
}
func (c *PaymentController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Payment, CreatePaymentDto, UpdatePaymentDto]{
		Name: "CapitalPayment", Searchable: []string{"user_id", "workspace_id", "plan_id", "reference"},
	})
}
