package message

import (
	"encoding/json"
	"time"

	"github.com/awesome-goose/goose/modules/sql"
)

type Message struct {
	sql.BaseEntity

	Reference   *string         `gorm:"column:reference;type:varchar(255)"    json:"reference,omitempty"`
	UserId      *string         `gorm:"column:user_id;type:varchar(36)"       json:"userId,omitempty"`
	ClientId    *string         `gorm:"column:client_id;type:varchar(255)"    json:"clientId,omitempty"`
	WorkspaceId *string         `gorm:"column:workspace_id;type:varchar(36)"  json:"workspaceId,omitempty"`
	Key         *string         `gorm:"column:key;type:varchar(255)"          json:"key,omitempty"`
	Subject     *string         `gorm:"column:subject;type:varchar(255)"      json:"subject,omitempty"`
	Channel     *string         `gorm:"column:channel;type:varchar(255)"      json:"channel,omitempty"`
	Message     json.RawMessage `gorm:"column:message;type:jsonb"             json:"message,omitempty"`
	Data        json.RawMessage `gorm:"column:data;type:jsonb"                json:"data,omitempty"`
	ReadAt      *time.Time      `gorm:"column:read_at;type:timestamp"         json:"readAt,omitempty"`
	PublishAt   *time.Time      `gorm:"column:publish_at;type:timestamp"      json:"publishAt,omitempty"`
	ExpireAt    *time.Time      `gorm:"column:expire_at;type:timestamp"       json:"expireAt,omitempty"`
	Type        *string         `gorm:"column:type;type:varchar(255)"         json:"type,omitempty"`
	Priority    *string         `gorm:"column:priority;type:varchar(255)"     json:"priority,omitempty"`
	Theme       *string         `gorm:"column:theme;type:varchar(255)"        json:"theme,omitempty"`
	Scope       *string         `gorm:"column:scope;type:varchar(255)"        json:"scope,omitempty"`
	Position    *string         `gorm:"column:position;type:varchar(255)"     json:"position,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"       json:"status,omitempty"`
}

func (Message) TableName() string { return "NotificationMessages" }

type MessageEntity struct {
	*sql.Entity[Message] `inject:""`
}

func (e *MessageEntity) OnRegister() {
	e.Hydrate("NotificationMessages", []string{"reference", "user_id", "client_id", "workspace_id", "key", "subject", "channel"}, nil, nil, nil, nil, nil, "created_at desc")
}
