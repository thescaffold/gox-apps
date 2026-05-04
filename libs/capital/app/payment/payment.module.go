package payment
import ("github.com/awesome-goose/goose/modules/sql"; "github.com/awesome-goose/goose/types")
type PaymentModule struct{}
func (m *PaymentModule) Imports() []types.Module { return []types.Module{ROUTES, sql.Child(&sql.Config{})} }
func (m *PaymentModule) Exports() []any { return []any{&PaymentService{}} }
func (m *PaymentModule) Declarations() []any { return []any{&PaymentService{}, &PaymentEntity{}} }
