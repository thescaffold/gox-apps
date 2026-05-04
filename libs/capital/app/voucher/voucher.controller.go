package voucher

import "github.com/thescaffold/gox-packages-core/crud"

type VoucherController struct {
	crud.CrudResource[Voucher, CreateVoucherDto, UpdateVoucherDto]
	entity *VoucherEntity `inject:""`
}

func (c *VoucherController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Voucher, CreateVoucherDto, UpdateVoucherDto]{
		Name:       "CapitalVoucher",
		Searchable: []string{"user_id", "client_id", "workspace_id", "voucher_type_id"},
	})
}
