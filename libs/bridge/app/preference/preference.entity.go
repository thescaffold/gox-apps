package preference

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Preference struct {
	sql.BaseEntity
	LicenseId string          `gorm:"column:license_id;type:varchar(36);not null" json:"licenseId"`
	Key       string          `gorm:"column:key;type:varchar(255);not null"       json:"key"`
	Value     *string         `gorm:"column:value;type:varchar(255)"              json:"value,omitempty"`
	Type      *string         `gorm:"column:type;type:varchar(255)"               json:"type,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}
func (Preference) TableName() string { return "BridgePreferences" }
type PreferenceEntity struct{ *sql.Entity[Preference] `inject:""` }
func (e *PreferenceEntity) OnRegister() {
	e.Hydrate("BridgePreferences", []string{"license_id", "key"}, nil, nil, nil, nil, nil, "created_at desc")
}
