package provider

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Provider struct {
	sql.BaseEntity
	UserId    string          `gorm:"column:user_id;type:varchar(36);not null"   json:"userId"`
	ClientId  *string         `gorm:"column:client_id;type:varchar(255)"         json:"clientId,omitempty"`
	Type      string          `gorm:"column:type;type:varchar(255);not null"     json:"type"`
	Reference string          `gorm:"column:reference;type:varchar(255);not null" json:"reference"`
	Email     *string         `gorm:"column:email;type:varchar(255)"             json:"email,omitempty"`
	Token     *string         `gorm:"column:token;type:text"                     json:"token,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                     json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (Provider) TableName() string { return "IdentityProviders" }

type ProviderEntity struct {
	*sql.Entity[Provider] `inject:""`
}

func (e *ProviderEntity) OnRegister() {
	e.Hydrate("IdentityProviders", []string{"user_id", "type", "reference"}, nil, nil, nil, nil, nil, "created_at desc")
}
