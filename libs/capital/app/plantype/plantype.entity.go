package plantype

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type PlanType struct {
	sql.BaseEntity
	ClientId string          `gorm:"column:client_id;type:varchar(255);not null" json:"clientId"`
	Key      string          `gorm:"column:key;type:varchar(255);not null"       json:"key"`
	Name     string          `gorm:"column:name;type:varchar(255);not null"      json:"name"`
	Desc     *string         `gorm:"column:desc;type:varchar(255)"               json:"desc,omitempty"`
	Detail   *string         `gorm:"column:detail;type:text"                     json:"detail,omitempty"`
	Type     *string         `gorm:"column:type;type:varchar(255)"               json:"type,omitempty"`
	Currency string          `gorm:"column:currency;type:varchar(255);not null"  json:"currency"`
	Daily    *int            `gorm:"column:daily"                                json:"daily,omitempty"`
	Weekly   *int            `gorm:"column:weekly"                               json:"weekly,omitempty"`
	Monthly  int             `gorm:"column:monthly;not null"                     json:"monthly"`
	Yearly   *int            `gorm:"column:yearly"                               json:"yearly,omitempty"`
	Meta     json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}
func (PlanType) TableName() string { return "CapitalPlanTypes" }
type PlanTypeEntity struct{ *sql.Entity[PlanType] `inject:""` }
func (e *PlanTypeEntity) OnRegister() {
	e.Hydrate("CapitalPlanTypes", []string{"client_id", "key", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
