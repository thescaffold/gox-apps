package token

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type Token struct {
	sql.BaseEntity

	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	Group       string          `gorm:"column:group;type:varchar(255);not null"       json:"group"`
	Service     string          `gorm:"column:service;type:varchar(255);not null"     json:"service"`
	EntityId    string          `gorm:"column:entity_id;type:varchar(36);not null"    json:"entityId"`
	EntityName  string          `gorm:"column:entity_name;type:varchar(255);not null" json:"entityName"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Token) TableName() string { return "FussTokens" }

type TokenEntity struct {
	*sql.Entity[Token] `inject:""`
}

func (e *TokenEntity) OnRegister() {
	e.Hydrate("FussTokens",
		[]string{"user_id", "client_id", "workspace_id", "group", "service", "entity_id", "entity_name"},
		nil, nil, nil, nil, nil, "created_at desc")
}
