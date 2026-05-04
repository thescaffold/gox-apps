package devicelog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type DeviceLog struct {
	sql.BaseEntity
	DeviceId string          `gorm:"column:device_id;type:varchar(36);not null" json:"deviceId"`
	Event    string          `gorm:"column:event;type:varchar(255);not null"    json:"event"`
	Meta     json.RawMessage `gorm:"column:meta;type:jsonb"                     json:"meta,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}
func (DeviceLog) TableName() string { return "IdentityDeviceLogs" }
type DeviceLogEntity struct{ *sql.Entity[DeviceLog] `inject:""` }
func (e *DeviceLogEntity) OnRegister() {
	e.Hydrate("IdentityDeviceLogs", []string{"device_id", "event"}, nil, nil, nil, nil, nil, "created_at desc")
}
