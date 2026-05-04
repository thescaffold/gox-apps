package provider

import "encoding/json"

type CreateProviderDto struct {
	AccountId *string         `json:"accountId,omitempty"`
	Name      string          `json:"name"      binding:"required"`
	Type      string          `json:"type"      binding:"required"`
	Signature *string         `json:"signature,omitempty"`
	Email     *string         `json:"email,omitempty"`
	Token     *string         `json:"token,omitempty"`
	Primary   *bool           `json:"primary,omitempty"`
	Currency  string          `json:"currency"  binding:"required"`
	Country   string          `json:"country"   binding:"required"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}

type UpdateProviderDto struct {
	AccountId *string         `json:"accountId,omitempty"`
	Name      *string         `json:"name,omitempty"`
	Type      *string         `json:"type,omitempty"`
	Signature *string         `json:"signature,omitempty"`
	Email     *string         `json:"email,omitempty"`
	Token     *string         `json:"token,omitempty"`
	Primary   *bool           `json:"primary,omitempty"`
	Currency  *string         `json:"currency,omitempty"`
	Country   *string         `json:"country,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}
