package webhook

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Webhook struct {
	sql.BaseEntity
	LicenseId string          `gorm:"column:license_id;type:varchar(36);not null" json:"licenseId"`
	Type      *string         `gorm:"column:type;type:varchar(255)"               json:"type,omitempty"`
	Url       *string         `gorm:"column:url;type:varchar(255)"                json:"url,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}

func (Webhook) TableName() string { return "BridgeWebhooks" }

type WebhookEntity struct {
	*sql.Entity[Webhook] `inject:""`
}

func (e *WebhookEntity) OnRegister() {
	e.Hydrate("BridgeWebhooks", []string{"license_id", "type"}, nil, nil, nil, nil, nil, "created_at desc")
}
