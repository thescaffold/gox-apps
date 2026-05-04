package transaction

import "time"

type CreateTransactionDto struct {
	UserId           string     `json:"userId"           binding:"required"`
	AccountId        string     `json:"accountId"        binding:"required"`
	Amount           int        `json:"amount"           binding:"required"`
	BookBalance      *int       `json:"bookBalance,omitempty"`
	AvailableBalance *int       `json:"availableBalance,omitempty"`
	Currency         string     `json:"currency"         binding:"required"`
	Reference        string     `json:"reference"        binding:"required"`
	Narration        *string    `json:"narration,omitempty"`
	Desc             *string    `json:"desc,omitempty"`
	Type             string     `json:"type"             binding:"required"`
	Status           *string    `json:"status,omitempty"`
	TransactionAt    *time.Time `json:"transactionAt,omitempty"`
	VerifiedAt       *time.Time `json:"verifiedAt,omitempty"`
	InvoicedAt       *time.Time `json:"invoicedAt,omitempty"`
	PaidAt           *time.Time `json:"paidAt,omitempty"`
}

type UpdateTransactionDto struct {
	Amount           *int       `json:"amount,omitempty"`
	BookBalance      *int       `json:"bookBalance,omitempty"`
	AvailableBalance *int       `json:"availableBalance,omitempty"`
	Narration        *string    `json:"narration,omitempty"`
	Desc             *string    `json:"desc,omitempty"`
	Type             *string    `json:"type,omitempty"`
	Status           *string    `json:"status,omitempty"`
	TransactionAt    *time.Time `json:"transactionAt,omitempty"`
	VerifiedAt       *time.Time `json:"verifiedAt,omitempty"`
	InvoicedAt       *time.Time `json:"invoicedAt,omitempty"`
	PaidAt           *time.Time `json:"paidAt,omitempty"`
}
