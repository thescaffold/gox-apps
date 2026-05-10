package payment

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
	"time"
)

type Payment struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	PlanId      string          `gorm:"column:plan_id;type:varchar(36);not null"      json:"planId"`
	PeriodType  *string         `gorm:"column:period_type;type:varchar(255)"          json:"periodType,omitempty"`
	Period      *string         `gorm:"column:period;type:varchar(255)"               json:"period,omitempty"`
	Amount      *int            `gorm:"column:amount"                                 json:"amount,omitempty"`
	Currency    *string         `gorm:"column:currency;type:varchar(255)"             json:"currency,omitempty"`
	Paid        *bool           `gorm:"column:paid;type:boolean"                      json:"paid,omitempty"`
	InvoiceUrl  *string         `gorm:"column:invoice_url;type:varchar(255)"          json:"invoiceUrl,omitempty"`
	ReceiptUrl  *string         `gorm:"column:receipt_url;type:varchar(255)"          json:"receiptUrl,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Type        string          `gorm:"column:type;type:varchar(255);not null"        json:"type"`
	Reference   string          `gorm:"column:reference;type:varchar(255);not null"   json:"reference"`
	Request     json.RawMessage `gorm:"column:request;type:jsonb"                     json:"request,omitempty"`
	Response    json.RawMessage `gorm:"column:response;type:jsonb"                    json:"response,omitempty"`
	StartAt     *time.Time      `gorm:"column:start_at;type:timestamp"                json:"startAt,omitempty"`
	EndAt       *time.Time      `gorm:"column:end_at;type:timestamp"                  json:"endAt,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Payment) TableName() string { return "CapitalPayments" }

type PaymentEntity struct {
	*sql.Entity[Payment] `inject:""`
}

func (e *PaymentEntity) OnRegister() {
	e.Hydrate("CapitalPayments", []string{"user_id", "workspace_id", "plan_id", "reference"}, nil, nil, nil, nil, nil, "created_at desc")
}
