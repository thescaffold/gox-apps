package provider

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Provider struct {
	sql.BaseEntity
	AccountId *string         `gorm:"column:account_id;type:varchar(36)"         json:"accountId,omitempty"`
	Name      string          `gorm:"column:name;type:varchar(255);not null"      json:"name"`
	Type      string          `gorm:"column:type;type:varchar(255);not null"      json:"type"`
	Signature *string         `gorm:"column:signature;type:varchar(255)"          json:"signature,omitempty"`
	Email     *string         `gorm:"column:email;type:varchar(255)"              json:"email,omitempty"`
	Token     *string         `gorm:"column:token;type:varchar(255)"              json:"token,omitempty"`
	Primary   *bool           `gorm:"column:primary;type:boolean"                 json:"primary,omitempty"`
	Currency  string          `gorm:"column:currency;type:varchar(255);not null"  json:"currency"`
	Country   string          `gorm:"column:country;type:varchar(255);not null"   json:"country"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}
func (Provider) TableName() string { return "CapitalProviders" }
type ProviderEntity struct{ *sql.Entity[Provider] `inject:""` }
func (e *ProviderEntity) OnRegister() {
	e.Hydrate("CapitalProviders", []string{"account_id", "name", "currency"}, nil, nil, nil, nil, nil, "created_at desc")
}
