package providerlog

import "encoding/json"

type CreateProviderLogDto struct {
	ProviderId string          `json:"providerId" binding:"required"`
	Event      string          `json:"event"      binding:"required"`
	Request    json.RawMessage `json:"request,omitempty"`
	Response   json.RawMessage `json:"response,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Status     *string         `json:"status,omitempty"`
}

type UpdateProviderLogDto struct {
	Response json.RawMessage `json:"response,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
