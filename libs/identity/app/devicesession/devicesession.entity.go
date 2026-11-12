package devicesession

import (
	"github.com/awesome-goose/goose/modules/sql"
	"time"
)

type DeviceSession struct {
	sql.BaseEntity
	DeviceId  string     `gorm:"column:device_id;type:varchar(36);not null" json:"deviceId"`
	UserId    string     `gorm:"column:user_id;type:varchar(36);not null"   json:"userId"`
	Token     string     `gorm:"column:token;type:text;not null"            json:"token"`
	ExpiresAt *time.Time `gorm:"column:expires_at;type:timestamp"           json:"expiresAt,omitempty"`
	Status    *string    `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
	// ClientId is the client the session signed in to (nil for sessions made
	// before PLAN M1-40).
	ClientId *string `gorm:"column:client_id;type:varchar(255)" json:"clientId,omitempty"`
	// Where it was last used from and when (V0-15, Account › Sessions).
	IP         string     `gorm:"column:ip;type:varchar(64);default:''"       json:"ip"`
	Place      string     `gorm:"column:place;type:varchar(255);default:''"   json:"place"`
	LastSeenAt *time.Time `gorm:"column:last_seen_at;type:timestamp"          json:"lastSeenAt,omitempty"`
	// A session opened through SAML (V4-03): what single logout names it by.
	SsoConnectionId string `gorm:"column:sso_connection_id;type:varchar(36);default:''" json:"-"`
	SsoNameId       string `gorm:"column:sso_name_id;type:text;default:''"              json:"-"`
	SsoSessionIndex string `gorm:"column:sso_session_index;type:text;default:''"        json:"-"`
}

func (DeviceSession) TableName() string { return "IdentityDeviceSessions" }

type DeviceSessionEntity struct {
	*sql.Entity[DeviceSession] `inject:""`
}

func (e *DeviceSessionEntity) OnRegister() {
	e.Hydrate("IdentityDeviceSessions", []string{"device_id", "user_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
