package environment

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type Environment struct {
	sql.BaseEntity

	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	TypeId      string          `gorm:"column:type_id;type:varchar(36);not null"      json:"typeId"`
	Name        string          `gorm:"column:name;type:varchar(255);not null"        json:"name"`
	Desc        *string         `gorm:"column:desc;type:varchar(255)"                 json:"desc,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Environment) TableName() string { return "FlagEnvironments" }

type EnvironmentEntity struct {
	*sql.Entity[Environment] `inject:""`
}

func (e *EnvironmentEntity) OnRegister() {
	// Mirrors TS environment.controller.ts:20 searchable = ['name','desc'].
	e.Hydrate("FlagEnvironments", []string{"name", "desc"}, nil, nil, nil, nil, nil, "created_at desc")
}
