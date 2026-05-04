package voucher

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type VoucherModule struct{}

func (m *VoucherModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *VoucherModule) Exports() []any { return []any{&VoucherService{}} }

func (m *VoucherModule) Declarations() []any {
	return []any{&VoucherService{}, &VoucherEntity{}}
}
