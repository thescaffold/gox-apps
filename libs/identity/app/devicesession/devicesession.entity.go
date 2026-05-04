package devicesession

import (
	"time"
	"github.com/awesome-goose/goose/modules/sql"
)

type DeviceSession struct {
	sql.BaseEntity
	DeviceId  string     `gorm:"column:device_id;type:varchar(36);not null" json:"deviceId"`
	UserId    string     `gorm:"column:user_id;type:varchar(36);not null"   json:"userId"`
	Token     string     `gorm:"column:token;type:text;not null"            json:"token"`
	ExpiresAt *time.Time `gorm:"column:expires_at;type:timestamp"           json:"expiresAt,omitempty"`
	Status    *string    `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}
func (DeviceSession) TableName() string { return "IdentityDeviceSessions" }
type DeviceSessionEntity struct{ *sql.Entity[DeviceSession] `inject:""` }
func (e *DeviceSessionEntity) OnRegister() {
	e.Hydrate("IdentityDeviceSessions", []string{"device_id", "user_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
