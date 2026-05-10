package paymentlog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type PaymentLogModule struct{}

func (m *PaymentLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *PaymentLogModule) Exports() []any { return []any{&PaymentLogService{}} }
func (m *PaymentLogModule) Declarations() []any {
	return []any{&PaymentLogService{}, &PaymentLogEntity{}}
}
