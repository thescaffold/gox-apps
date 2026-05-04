package rate

import "github.com/awesome-goose/goose/modules/sql"

type Rate struct {
	sql.BaseEntity

	Currency string  `gorm:"column:currency;type:varchar(255);not null;uniqueIndex:idx_common_rates_currency" json:"currency"`
	Value    *int    `gorm:"column:value;type:integer"     json:"value,omitempty"`
	Delta    *int    `gorm:"column:delta;type:integer"     json:"delta,omitempty"`
	Status   *string `gorm:"column:status;type:varchar(255)" json:"status,omitempty"`
}

func (Rate) TableName() string { return "CommonRates" }

type RateEntity struct {
	*sql.Entity[Rate] `inject:""`
}

func (e *RateEntity) OnRegister() {
	e.Hydrate("CommonRates", []string{"currency"}, nil, nil, nil, nil, nil, "created_at desc")
}
