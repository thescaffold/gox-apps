package devicelog

import "encoding/json"

type CreateDeviceLogDto struct {
	DeviceId string          `json:"deviceId" binding:"required"`
	Event    string          `json:"event"    binding:"required"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

type UpdateDeviceLogDto struct {
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
