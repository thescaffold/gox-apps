package devicelog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

// DeviceLog mirrors ntx-apps/libs/identity/src/api/device-log/entities/device-log.entity.ts.
// TS carries request telemetry: session_id + ip/country/region/city/area. gox
// previously only had device_id+event+meta — Phase 4 adds the TS columns
// alongside the gox columns.
type DeviceLog struct {
	sql.BaseEntity
	DeviceId  string          `gorm:"column:device_id;type:varchar(36);not null" json:"deviceId"`
	SessionId *string         `gorm:"column:session_id;type:varchar(36)"         json:"sessionId,omitempty"`
	Ip        *string         `gorm:"column:ip;type:varchar(255)"                json:"ip,omitempty"`
	Country   *string         `gorm:"column:country;type:varchar(255)"           json:"country,omitempty"`
	Region    *string         `gorm:"column:region;type:varchar(255)"            json:"region,omitempty"`
	City      *string         `gorm:"column:city;type:varchar(255)"              json:"city,omitempty"`
	Area      *string         `gorm:"column:area;type:varchar(255)"              json:"area,omitempty"`
	Event     string          `gorm:"column:event;type:varchar(255);not null"    json:"event"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                     json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (DeviceLog) TableName() string { return "IdentityDeviceLogs" }

type DeviceLogEntity struct {
	*sql.Entity[DeviceLog] `inject:""`
}

func (e *DeviceLogEntity) OnRegister() {
	e.Hydrate("IdentityDeviceLogs", []string{"device_id", "event"}, nil, nil, nil, nil, nil, "created_at desc")
}
