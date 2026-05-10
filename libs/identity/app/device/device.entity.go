package device

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Device struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	Fingerprint string          `gorm:"column:fingerprint;type:varchar(255);not null" json:"fingerprint"`
	Type        string          `gorm:"column:type;type:varchar(255);not null"        json:"type"`
	Platform    *string         `gorm:"column:platform;type:varchar(255)"             json:"platform,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Device) TableName() string { return "IdentityDevices" }

type DeviceEntity struct {
	*sql.Entity[Device] `inject:""`
}

func (e *DeviceEntity) OnRegister() {
	e.Hydrate("IdentityDevices", []string{"user_id", "fingerprint"}, nil, nil, nil, nil, nil, "created_at desc")
}
