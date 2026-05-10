package paymentlog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type PaymentLog struct {
	sql.BaseEntity
	PaymentId string          `gorm:"column:payment_id;type:varchar(36);not null" json:"paymentId"`
	Type      string          `gorm:"column:type;type:varchar(255);not null"      json:"type"`
	Request   json.RawMessage `gorm:"column:request;type:jsonb;not null"          json:"request"`
	Response  json.RawMessage `gorm:"column:response;type:jsonb"                  json:"response,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}

func (PaymentLog) TableName() string { return "CapitalPaymentLogs" }

type PaymentLogEntity struct {
	*sql.Entity[PaymentLog] `inject:""`
}

func (e *PaymentLogEntity) OnRegister() {
	e.Hydrate("CapitalPaymentLogs", []string{"payment_id", "type"}, nil, nil, nil, nil, nil, "created_at desc")
}
