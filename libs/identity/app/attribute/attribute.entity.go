package attribute

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Attribute struct {
	sql.BaseEntity

	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"   json:"userId"`
	ClientId    *string         `gorm:"column:client_id;type:varchar(255)"         json:"clientId,omitempty"`
	WorkspaceId *string         `gorm:"column:workspace_id;type:varchar(36)"       json:"workspaceId,omitempty"`
	Key         string          `gorm:"column:key;type:varchar(255);not null"      json:"key"`
	Value       *string         `gorm:"column:value;type:varchar(255)"             json:"value,omitempty"`
	Type        *string         `gorm:"column:type;type:varchar(255)"              json:"type,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                     json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (Attribute) TableName() string { return "IdentityAttributes" }

type AttributeEntity struct {
	*sql.Entity[Attribute] `inject:""`
}

func (e *AttributeEntity) OnRegister() {
	e.Hydrate("IdentityAttributes", []string{"user_id", "client_id", "workspace_id", "key"}, nil, nil, nil, nil, nil, "created_at desc")
}
