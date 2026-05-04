package ratelog

import (
	"time"

	"github.com/awesome-goose/goose/modules/sql"
)

type RateLog struct {
	sql.BaseEntity

	RateId string     `gorm:"column:rate_id;type:varchar(36);not null" json:"rateId"`
	Value  *int       `gorm:"column:value;type:integer"               json:"value,omitempty"`
	Delta  *int       `gorm:"column:delta;type:integer"               json:"delta,omitempty"`
	Date   *time.Time `gorm:"column:date;type:timestamp"              json:"date,omitempty"`
	Status *string    `gorm:"column:status;type:varchar(255)"         json:"status,omitempty"`
}

func (RateLog) TableName() string { return "CommonRateLogs" }

type RateLogEntity struct {
	*sql.Entity[RateLog] `inject:""`
}

func (e *RateLogEntity) OnRegister() {
	e.Hydrate("CommonRateLogs", []string{"rate_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
