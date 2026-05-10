package client

import "github.com/awesome-goose/goose/modules/sql"

type Client struct {
	sql.BaseEntity
	Name   string  `gorm:"column:name;type:varchar(255);not null"  json:"name"`
	Type   string  `gorm:"column:type;type:varchar(255);not null"  json:"type"`
	Key    string  `gorm:"column:key;type:varchar(255);not null"   json:"key"`
	Secret *string `gorm:"column:secret;type:text"                 json:"-"`
	Desc   *string `gorm:"column:desc;type:varchar(255)"           json:"desc,omitempty"`
	Status *string `gorm:"column:status;type:varchar(255)"         json:"status,omitempty"`
}

func (Client) TableName() string { return "IdentityClients" }

type ClientEntity struct {
	*sql.Entity[Client] `inject:""`
}

func (e *ClientEntity) OnRegister() {
	e.Hydrate("IdentityClients", []string{"key", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
