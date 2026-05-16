package summary

import "github.com/awesome-goose/goose/modules/sql"

type Summary struct {
	sql.BaseEntity

	ServiceId string  `gorm:"column:service_id;type:varchar(36);not null" json:"serviceId"`
	Type      string  `gorm:"column:type;type:varchar(255);not null"      json:"type"`
	Received  int     `gorm:"column:received;not null;default:0"          json:"received"`
	Measure   float64 `gorm:"column:measure;not null"                     json:"measure"`
	Note      *string `gorm:"column:note;type:varchar(500)"               json:"note,omitempty"`
}

func (Summary) TableName() string { return "HealthSummaries" }

type SummaryEntity struct {
	*sql.Entity[Summary] `inject:""`
}

func (e *SummaryEntity) OnRegister() {
	// Mirrors TS summary.controller.ts:19 searchable = ['note'].
	e.Hydrate("HealthSummaries", []string{"note"}, nil, nil, nil, nil, nil, "created_at desc")
}
