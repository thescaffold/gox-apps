package plan

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Plan struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	TypeId      string          `gorm:"column:type_id;type:varchar(36);not null"      json:"typeId"`
	PeriodType  *string         `gorm:"column:period_type;type:varchar(255)"          json:"periodType,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Plan) TableName() string { return "CapitalPlans" }

type PlanEntity struct {
	*sql.Entity[Plan] `inject:""`
}

func (e *PlanEntity) OnRegister() {
	e.Hydrate("CapitalPlans", []string{"user_id", "workspace_id", "type_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
