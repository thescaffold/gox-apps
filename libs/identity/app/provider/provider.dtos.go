package provider

import "encoding/json"

type CreateProviderDto struct {
	UserId    string          `json:"userId"     binding:"required"`
	ClientId  *string         `json:"clientId,omitempty"`
	Type      string          `json:"type"       binding:"required"`
	Reference string          `json:"reference"  binding:"required"`
	Email     *string         `json:"email,omitempty"`
	Token     *string         `json:"token,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}

type UpdateProviderDto struct {
	Email  *string         `json:"email,omitempty"`
	Token  *string         `json:"token,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
