package log

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type LogActionType = string

const (
	LogActionTypeCreate LogActionType = "create"
	LogActionTypeRead   LogActionType = "read"
	LogActionTypeUpdate LogActionType = "update"
	LogActionTypeDelete LogActionType = "delete"
)

type Log struct {
	sql.BaseEntity

	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	Group       string          `gorm:"column:group;type:varchar(255);not null"       json:"group"`
	Service     string          `gorm:"column:service;type:varchar(255);not null"     json:"service"`
	EntityId    string          `gorm:"column:entity_id;type:varchar(36);not null"    json:"entityId"`
	EntityName  string          `gorm:"column:entity_name;type:varchar(255);not null" json:"entityName"`
	Action      string          `gorm:"column:action;type:varchar(255);not null"      json:"action"`
	Desc        *string         `gorm:"column:desc;type:varchar(255)"                 json:"desc,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Log) TableName() string { return "AuditLogs" }

type LogEntity struct {
	*sql.Entity[Log] `inject:""`
}

func (e *LogEntity) OnRegister() {
	e.Hydrate(
		"AuditLogs",
		[]string{"desc"}, // mirrors TS log.controller.ts searchable
		[]string{},
		nil,
		nil,
		nil,
		nil,
		"created_at desc",
	)
}
