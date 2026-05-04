package log

import (
	"encoding/json"
	"time"

	"github.com/awesome-goose/goose/modules/sql"
)

type Log struct {
	sql.BaseEntity

	Reference   string          `gorm:"column:reference;type:varchar(255);not null"  json:"reference"`
	UserId      *string         `gorm:"column:user_id;type:varchar(36)"              json:"userId,omitempty"`
	ClientId    *string         `gorm:"column:client_id;type:varchar(255)"           json:"clientId,omitempty"`
	WorkspaceId *string         `gorm:"column:workspace_id;type:varchar(36)"         json:"workspaceId,omitempty"`
	Key         string          `gorm:"column:key;type:varchar(255);not null"        json:"key"`
	Subject     string          `gorm:"column:subject;type:varchar(255);not null"    json:"subject"`
	Channels    json.RawMessage `gorm:"column:channels;type:jsonb;not null"          json:"channels"`
	Data        json.RawMessage `gorm:"column:data;type:jsonb"                       json:"data,omitempty"`
	Message     json.RawMessage `gorm:"column:message;type:jsonb"                    json:"message,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                       json:"meta,omitempty"`
	ReadAt      *time.Time      `gorm:"column:read_at;type:timestamp"                json:"readAt,omitempty"`
	PublishAt   *time.Time      `gorm:"column:publish_at;type:timestamp"             json:"publishAt,omitempty"`
	ExpireAt    *time.Time      `gorm:"column:expire_at;type:timestamp"              json:"expireAt,omitempty"`
	Type        *string         `gorm:"column:type;type:varchar(255)"                json:"type,omitempty"`
	Priority    *string         `gorm:"column:priority;type:varchar(255)"            json:"priority,omitempty"`
	Theme       *string         `gorm:"column:theme;type:varchar(255)"               json:"theme,omitempty"`
	Scope       *string         `gorm:"column:scope;type:varchar(255)"               json:"scope,omitempty"`
	Position    *string         `gorm:"column:position;type:varchar(255)"            json:"position,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"              json:"status,omitempty"`
	TemplateId  *string         `gorm:"column:template_id;type:varchar(36)"          json:"templateId,omitempty"`
	RuleId      *string         `gorm:"column:rule_id;type:varchar(36)"              json:"ruleId,omitempty"`
}

func (Log) TableName() string { return "NotificationLogs" }

type LogEntity struct {
	*sql.Entity[Log] `inject:""`
}

func (e *LogEntity) OnRegister() {
	e.Hydrate("NotificationLogs", []string{"reference", "user_id", "workspace_id", "key"}, nil, nil, nil, nil, nil, "created_at desc")
}
