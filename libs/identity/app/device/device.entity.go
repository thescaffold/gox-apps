package device

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

// Device mirrors ntx-apps/libs/identity/src/api/device/entities/device.entity.ts.
// TS describes the browser/User-Agent components (os/agent/engine/cpu) while
// gox originally stored a per-user fingerprint (fingerprint/type/platform).
// Phase 4 adds the TS columns alongside the gox columns — both populated by
// the appropriate code path.
type Device struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	Os          *string         `gorm:"column:os;type:varchar(255)"                   json:"os,omitempty"`
	Agent       *string         `gorm:"column:agent;type:varchar(255)"                json:"agent,omitempty"`
	Engine      *string         `gorm:"column:engine;type:varchar(255)"               json:"engine,omitempty"`
	Cpu         *string         `gorm:"column:cpu;type:varchar(255)"                  json:"cpu,omitempty"`
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
