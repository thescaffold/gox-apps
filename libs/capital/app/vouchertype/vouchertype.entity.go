package vouchertype

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type VoucherType struct {
	sql.BaseEntity
	Name     string          `gorm:"column:name;type:varchar(255);not null"     json:"name"`
	Desc     *string         `gorm:"column:desc;type:varchar(255)"              json:"desc,omitempty"`
	Token    string          `gorm:"column:token;type:varchar(255);not null"    json:"token"`
	Amount   int             `gorm:"column:amount;not null"                     json:"amount"`
	Currency string          `gorm:"column:currency;type:varchar(255);not null" json:"currency"`
	Rules    json.RawMessage `gorm:"column:rules;type:jsonb"                    json:"rules,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}
func (VoucherType) TableName() string { return "CapitalVoucherTypes" }
type VoucherTypeEntity struct{ *sql.Entity[VoucherType] `inject:""` }
func (e *VoucherTypeEntity) OnRegister() {
	e.Hydrate("CapitalVoucherTypes", []string{"name", "token"}, nil, nil, nil, nil, nil, "created_at desc")
}
