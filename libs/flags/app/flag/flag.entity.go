package flag

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type Flag struct {
	sql.BaseEntity

	UserId        string          `gorm:"column:user_id;type:varchar(36);not null"        json:"userId"`
	ClientId      string          `gorm:"column:client_id;type:varchar(255);not null"     json:"clientId"`
	WorkspaceId   string          `gorm:"column:workspace_id;type:varchar(36);not null"   json:"workspaceId"`
	EnvironmentId string          `gorm:"column:environment_id;type:varchar(36);not null" json:"environmentId"`
	Name          string          `gorm:"column:name;type:varchar(255);not null"          json:"name"`
	Limit         int             `gorm:"column:limit;not null"                           json:"limit"`
	Priority      int             `gorm:"column:priority;not null"                        json:"priority"`
	Level         string          `gorm:"column:level;type:varchar(255);not null"         json:"level"`
	Meta          json.RawMessage `gorm:"column:meta;type:jsonb"                          json:"meta,omitempty"`
	Status        *string         `gorm:"column:status;type:varchar(255)"                 json:"status,omitempty"`
}

func (Flag) TableName() string { return "FlagFlags" }

type FlagEntity struct {
	*sql.Entity[Flag] `inject:""`
}

func (e *FlagEntity) OnRegister() {
	e.Hydrate("FlagFlags", []string{"name", "workspace_id", "environment_id", "level"}, nil, nil, nil, nil, nil, "created_at desc")
}
