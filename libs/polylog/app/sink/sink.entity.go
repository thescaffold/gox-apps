package sink

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Sink struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	Category    string          `gorm:"column:category;type:varchar(255);not null"    json:"category"`
	Key         *string         `gorm:"column:key;type:varchar(255)"                  json:"key,omitempty"`
	Visibility  *string         `gorm:"column:visibility;type:varchar(255)"           json:"visibility,omitempty"`
	TypeId      string          `gorm:"column:type_id;type:varchar(36);not null"      json:"typeId"`
	Name        string          `gorm:"column:name;type:varchar(255);not null"        json:"name"`
	Desc        *string         `gorm:"column:desc;type:varchar(255)"                 json:"desc,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}
func (Sink) TableName() string { return "PolylogSinks" }
type SinkEntity struct{ *sql.Entity[Sink] `inject:""` }
func (e *SinkEntity) OnRegister() {
	e.Hydrate("PolylogSinks", []string{"user_id", "workspace_id", "category", "type_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
