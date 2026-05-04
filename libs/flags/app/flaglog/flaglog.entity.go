package flaglog

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type FlagLog struct {
	sql.BaseEntity

	FlagId string          `gorm:"column:flag_id;type:varchar(36);not null" json:"flagId"`
	Limit  int             `gorm:"column:limit;not null"                    json:"limit"`
	Meta   json.RawMessage `gorm:"column:meta;type:jsonb"                   json:"meta,omitempty"`
	Status *string         `gorm:"column:status;type:varchar(255)"          json:"status,omitempty"`
}

func (FlagLog) TableName() string { return "FlagFlagLogs" }

type FlagLogEntity struct {
	*sql.Entity[FlagLog] `inject:""`
}

func (e *FlagLogEntity) OnRegister() {
	e.Hydrate("FlagFlagLogs", []string{"flag_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
