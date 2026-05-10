package webhooklog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type WebhookLog struct {
	sql.BaseEntity
	WebhookId string          `gorm:"column:webhook_id;type:varchar(36);not null" json:"webhookId"`
	Type      *string         `gorm:"column:type;type:varchar(255)"               json:"type,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Request   json.RawMessage `gorm:"column:request;type:jsonb"                   json:"request,omitempty"`
	Response  json.RawMessage `gorm:"column:response;type:jsonb"                  json:"response,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}

func (WebhookLog) TableName() string { return "BridgeWebhookLogs" }

type WebhookLogEntity struct {
	*sql.Entity[WebhookLog] `inject:""`
}

func (e *WebhookLogEntity) OnRegister() {
	e.Hydrate("BridgeWebhookLogs", []string{"webhook_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
