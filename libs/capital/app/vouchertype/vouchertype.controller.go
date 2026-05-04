package vouchertype

import "github.com/thescaffold/gox-packages-core/crud"

type VoucherTypeController struct {
	crud.CrudResource[VoucherType, CreateVoucherTypeDto, UpdateVoucherTypeDto]
	entity *VoucherTypeEntity `inject:""`
}

func (c *VoucherTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[VoucherType, CreateVoucherTypeDto, UpdateVoucherTypeDto]{
		Name:       "CapitalVoucherType",
		Searchable: []string{"name", "token"},
	})
}
