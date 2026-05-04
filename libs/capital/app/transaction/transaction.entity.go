package transaction

import (
	"time"
	"github.com/awesome-goose/goose/modules/sql"
)

type Transaction struct {
	sql.BaseEntity
	UserId           string     `gorm:"column:user_id;type:varchar(36);not null"       json:"userId"`
	AccountId        string     `gorm:"column:account_id;type:varchar(36);not null"    json:"accountId"`
	Amount           int        `gorm:"column:amount;not null"                         json:"amount"`
	BookBalance      int        `gorm:"column:book_balance;not null"                   json:"bookBalance"`
	AvailableBalance int        `gorm:"column:available_balance;not null"              json:"availableBalance"`
	Currency         string     `gorm:"column:currency;type:varchar(255);not null"     json:"currency"`
	Reference        string     `gorm:"column:reference;type:varchar(255);not null"    json:"reference"`
	Narration        *string    `gorm:"column:narration;type:varchar(255)"             json:"narration,omitempty"`
	Desc             *string    `gorm:"column:desc;type:text"                          json:"desc,omitempty"`
	Type             string     `gorm:"column:type;type:varchar(255);not null"         json:"type"`
	Status           *string    `gorm:"column:status;type:varchar(255)"                json:"status,omitempty"`
	TransactionAt    time.Time  `gorm:"column:transaction_at;type:timestamp;not null"  json:"transactionAt"`
	VerifiedAt       *time.Time `gorm:"column:verified_at;type:timestamp"              json:"verifiedAt,omitempty"`
	InvoicedAt       *time.Time `gorm:"column:invoiced_at;type:timestamp"              json:"invoicedAt,omitempty"`
	PaidAt           *time.Time `gorm:"column:paid_at;type:timestamp"                  json:"paidAt,omitempty"`
}
func (Transaction) TableName() string { return "CapitalTransactions" }
type TransactionEntity struct{ *sql.Entity[Transaction] `inject:""` }
func (e *TransactionEntity) OnRegister() {
	e.Hydrate("CapitalTransactions", []string{"user_id", "account_id", "currency", "reference"}, nil, nil, nil, nil, nil, "created_at desc")
}
