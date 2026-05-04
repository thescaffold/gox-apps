package licensetype

import "github.com/awesome-goose/goose/modules/sql"

type LicenseType struct {
	sql.BaseEntity

	Key      string   `gorm:"column:key;type:varchar(255);not null"         json:"key"`
	Name     string   `gorm:"column:name;type:varchar(255);not null"        json:"name"`
	Desc     *string  `gorm:"column:desc;type:varchar(255)"                json:"desc,omitempty"`
	Detail   *string  `gorm:"column:detail;type:text"                      json:"detail,omitempty"`
	Type     *string  `gorm:"column:type;type:varchar(255)"                json:"type,omitempty"`
	Currency string   `gorm:"column:currency;type:varchar(255);not null"   json:"currency"`
	Daily    *float64 `gorm:"column:daily;type:decimal(18,2)"              json:"daily,omitempty"`
	Weekly   *float64 `gorm:"column:weekly;type:decimal(18,2)"             json:"weekly,omitempty"`
	Monthly  float64  `gorm:"column:monthly;type:decimal(18,2);not null"   json:"monthly"`
	Yearly   *float64 `gorm:"column:yearly;type:decimal(18,2)"             json:"yearly,omitempty"`
	Meta     *string  `gorm:"column:meta;type:jsonb"                       json:"meta,omitempty"`
	Status   *string  `gorm:"column:status;type:varchar(255)"              json:"status,omitempty"`
}

func (LicenseType) TableName() string { return "BridgeLicenseTypes" }

type LicenseTypeEntity struct {
	*sql.Entity[LicenseType] `inject:""`
}

func (e *LicenseTypeEntity) OnRegister() {
	e.Hydrate("BridgeLicenseTypes", []string{"key", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
