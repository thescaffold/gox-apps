package log

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type Log struct {
	sql.BaseEntity

	ServiceId string          `gorm:"column:service_id;type:varchar(36);not null" json:"serviceId"`
	State     *string         `gorm:"column:state;type:varchar(255)"              json:"state,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
}

func (Log) TableName() string { return "HealthLogs" }

type LogEntity struct {
	*sql.Entity[Log] `inject:""`
}

func (e *LogEntity) OnRegister() {
	// Mirrors TS log.controller.ts:19 searchable = [] — Log is not query-filtered.
	e.Hydrate("HealthLogs", nil, nil, nil, nil, nil, nil, "created_at desc")
}
