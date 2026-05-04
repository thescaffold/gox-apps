package license

import (
	"time"

	"github.com/awesome-goose/goose/modules/sql"
)

type License struct {
	sql.BaseEntity

	UserId      string     `gorm:"column:user_id;type:varchar(36)"      json:"userId"`
	ClientId    string     `gorm:"column:client_id;type:varchar(255)"   json:"clientId"`
	WorkspaceId string     `gorm:"column:workspace_id;type:varchar(36)" json:"workspaceId"`
	TypeId      string     `gorm:"column:type_id;type:varchar(36)"      json:"typeId"`
	PeriodType  *string    `gorm:"column:period_type;type:varchar(255)" json:"periodType,omitempty"`
	Token       *string    `gorm:"column:token;type:text"               json:"token,omitempty"`
	StartAt     *time.Time `gorm:"column:start_at;type:timestamp"       json:"startAt,omitempty"`
	RenewedAt   *time.Time `gorm:"column:renewed_at;type:timestamp"     json:"renewedAt,omitempty"`
	ExpiredAt   *time.Time `gorm:"column:expired_at;type:timestamp"     json:"expiredAt,omitempty"`
	Meta        *string    `gorm:"column:meta;type:jsonb"               json:"meta,omitempty"`
	Status      *string    `gorm:"column:status;type:varchar(255)"      json:"status,omitempty"`
}

func (License) TableName() string { return "BridgeLicenses" }

type LicenseEntity struct {
	*sql.Entity[License] `inject:""`
}

func (e *LicenseEntity) OnRegister() {
	e.Hydrate("BridgeLicenses", []string{"user_id", "workspace_id", "type_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
