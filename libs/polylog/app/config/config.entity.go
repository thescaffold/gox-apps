package config

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

// Config persists per-source/channel/event polylog config rows.
// Mirrors ntx-apps/libs/polylog/src/api/config (TypeORM Config entity).
//
// The unique tuple is (userId, clientId, workspaceId, entityId, entityName,
// type, category) — looked up by GetConfig and used by SetConfig as the
// upsert key.
type Config struct {
	sql.BaseEntity

	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"        json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"     json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null"   json:"workspaceId"`
	EntityId    string          `gorm:"column:entity_id;type:varchar(36);not null"      json:"entityId"`
	EntityName  string          `gorm:"column:entity_name;type:varchar(255);not null"   json:"entityName"`
	Type        string          `gorm:"column:type;type:varchar(255);not null"          json:"type"`
	Category    string          `gorm:"column:category;type:varchar(255);not null"      json:"category"`
	Value       json.RawMessage `gorm:"column:value;type:jsonb"                         json:"value,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"                 json:"status,omitempty"`
}

func (Config) TableName() string { return "PolylogConfigs" }

type ConfigEntity struct {
	*sql.Entity[Config] `inject:""`
}

func (e *ConfigEntity) OnRegister() {
	e.Hydrate(
		"PolylogConfigs",
		[]string{"entity_name", "type", "category"},
		nil, nil, nil, nil, nil,
		"created_at desc",
	)
}
