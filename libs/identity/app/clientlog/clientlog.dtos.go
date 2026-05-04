package clientlog

import "encoding/json"

type CreateClientLogDto struct {
	ClientId string          `json:"clientId" binding:"required"`
	Event    string          `json:"event"    binding:"required"`
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

type UpdateClientLogDto struct {
	Response json.RawMessage `json:"response,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
