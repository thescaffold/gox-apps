package usage

import (
	"encoding/json"
	"time"
	"github.com/awesome-goose/goose/modules/sql"
)

type Usage struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	RateId      string          `gorm:"column:rate_id;type:varchar(36);not null"      json:"rateId"`
	EntityId    string          `gorm:"column:entity_id;type:varchar(36);not null"    json:"entityId"`
	EntityName  string          `gorm:"column:entity_name;type:varchar(255);not null" json:"entityName"`
	Quantity    int             `gorm:"column:quantity;not null"                      json:"quantity"`
	StartAt     *time.Time      `gorm:"column:start_at;type:timestamp"                json:"startAt,omitempty"`
	StopAt      *time.Time      `gorm:"column:stop_at;type:timestamp"                 json:"stopAt,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}
func (Usage) TableName() string { return "CapitalUsages" }
type UsageEntity struct{ *sql.Entity[Usage] `inject:""` }
func (e *UsageEntity) OnRegister() {
	e.Hydrate("CapitalUsages", []string{"user_id", "client_id", "workspace_id", "rate_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
