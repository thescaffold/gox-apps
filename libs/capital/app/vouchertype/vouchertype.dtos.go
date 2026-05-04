package vouchertype

import "encoding/json"

type CreateVoucherTypeDto struct {
	Name     string          `json:"name"     binding:"required"`
	Desc     *string         `json:"desc,omitempty"`
	Token    string          `json:"token"    binding:"required"`
	Amount   int             `json:"amount"   binding:"required"`
	Currency string          `json:"currency" binding:"required"`
	Rules    json.RawMessage `json:"rules,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

type UpdateVoucherTypeDto struct {
	Name     *string         `json:"name,omitempty"`
	Desc     *string         `json:"desc,omitempty"`
	Token    *string         `json:"token,omitempty"`
	Amount   *int            `json:"amount,omitempty"`
	Currency *string         `json:"currency,omitempty"`
	Rules    json.RawMessage `json:"rules,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
