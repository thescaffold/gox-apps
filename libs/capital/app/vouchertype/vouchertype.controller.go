package vouchertype

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// VoucherTypeController mirrors ntx-apps/libs/capital/src/api/voucher-type/voucher-type.controller.ts.
type VoucherTypeController struct {
	crud.CrudResource[VoucherType, CreateVoucherTypeDto, UpdateVoucherTypeDto]

	entity *VoucherTypeEntity `inject:""`
	lang   *i18n.Service      `inject:""`
}

func (c *VoucherTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[VoucherType, CreateVoucherTypeDto, UpdateVoucherTypeDto]{
		// Mirrors TS voucher-type.controller.ts name = 'voucher group'.
		Name: "voucher group",
		// Mirrors TS voucher-type.controller.ts searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS voucher-type.controller.ts unique = ({name,token}) =>
		// [{name},{token}] — two OR'd constraints.
		Unique: func(d *CreateVoucherTypeDto) []map[string]any {
			return []map[string]any{
				{"name": d.Name},
				{"token": d.Token},
			}
		},
	})
	c.SetLang(c.lang)
}
