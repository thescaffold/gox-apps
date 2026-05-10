package rate

import "github.com/awesome-goose/goose/modules/sql"

type Rate struct {
	sql.BaseEntity
	Type     string  `gorm:"column:type;type:varchar(255);not null"     json:"type"`
	PerUnit  *int    `gorm:"column:per_unit"                            json:"perUnit,omitempty"`
	Currency string  `gorm:"column:currency;type:varchar(255);not null" json:"currency"`
	Status   *string `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (Rate) TableName() string { return "CapitalRates" }

type RateEntity struct {
	*sql.Entity[Rate] `inject:""`
}

func (e *RateEntity) OnRegister() {
	e.Hydrate("CapitalRates", []string{"type", "currency"}, nil, nil, nil, nil, nil, "created_at desc")
}
