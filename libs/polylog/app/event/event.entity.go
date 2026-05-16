package event

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

// Event mirrors ntx-apps/libs/polylog/src/api/event/entities/event.entity.ts.
// TS stores the entity-name under `entity` (TS field `entityName` with
// `@Column({name: 'entity'})`); gox originally used `entity_name`. Phase 4
// adds the `entity` column; both columns coexist until a follow-up cleanup
// migration retires `entity_name`.
type Event struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	EntityId    string          `gorm:"column:entity_id;type:varchar(36);not null"    json:"entityId"`
	Entity      *string         `gorm:"column:entity;type:varchar(255)"               json:"-"`
	EntityName  string          `gorm:"column:entity_name;type:varchar(255);not null" json:"entityName"`
	Category    string          `gorm:"column:category;type:varchar(255);not null"    json:"category"`
	Type        *string         `gorm:"column:type;type:varchar(255)"                 json:"type,omitempty"`
	Reference   string          `gorm:"column:reference;type:varchar(255);not null"   json:"reference"`
	Version     string          `gorm:"column:version;type:varchar(255);not null"     json:"version"`
	Payload     json.RawMessage `gorm:"column:payload;type:jsonb;not null"            json:"payload"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Event) TableName() string { return "PolylogEvents" }

type EventEntity struct {
	*sql.Entity[Event] `inject:""`
}

func (e *EventEntity) OnRegister() {
	e.Hydrate("PolylogEvents", []string{"user_id", "workspace_id", "entity_name", "category", "reference"}, nil, nil, nil, nil, nil, "created_at desc")
}
