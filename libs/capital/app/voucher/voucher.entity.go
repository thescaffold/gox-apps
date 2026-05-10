package voucher

import "github.com/awesome-goose/goose/modules/sql"

type Voucher struct {
	sql.BaseEntity
	UserId        string  `gorm:"column:user_id;type:varchar(36);not null"        json:"userId"`
	ClientId      string  `gorm:"column:client_id;type:varchar(255);not null"     json:"clientId"`
	WorkspaceId   string  `gorm:"column:workspace_id;type:varchar(36);not null"   json:"workspaceId"`
	VoucherTypeId string  `gorm:"column:voucher_type_id;type:varchar(36);not null" json:"voucherTypeId"`
	Status        *string `gorm:"column:status;type:varchar(255)"                 json:"status,omitempty"`
}

func (Voucher) TableName() string { return "CapitalVouchers" }

type VoucherEntity struct {
	*sql.Entity[Voucher] `inject:""`
}

func (e *VoucherEntity) OnRegister() {
	e.Hydrate("CapitalVouchers", []string{"user_id", "client_id", "workspace_id", "voucher_type_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
