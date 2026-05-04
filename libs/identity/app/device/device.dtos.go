package device

import "encoding/json"

type CreateDeviceDto struct {
	UserId      string          `json:"userId"      binding:"required"`
	Fingerprint string          `json:"fingerprint" binding:"required"`
	Type        string          `json:"type"        binding:"required"`
	Platform    *string         `json:"platform,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateDeviceDto struct {
	Type     *string         `json:"type,omitempty"`
	Platform *string         `json:"platform,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
