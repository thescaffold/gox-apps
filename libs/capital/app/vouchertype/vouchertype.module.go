package vouchertype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type VoucherTypeModule struct{}

func (m *VoucherTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *VoucherTypeModule) Exports() []any { return []any{&VoucherTypeService{}} }

func (m *VoucherTypeModule) Declarations() []any {
	return []any{&VoucherTypeService{}, &VoucherTypeEntity{}}
}
