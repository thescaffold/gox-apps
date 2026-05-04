package log

import "encoding/json"

type CreateLogDto struct {
	ServiceId string          `json:"serviceId" binding:"required"`
	State     *string         `json:"state,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
}

type UpdateLogDto struct {
	ServiceId *string         `json:"serviceId,omitempty"`
	State     *string         `json:"state,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
}
