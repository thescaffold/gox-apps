package list

import (
	"encoding/json"
	"time"
)

type CreateListDto struct {
	Key       string          `json:"key"                binding:"required"`
	Value     json.RawMessage `json:"value,omitempty"`
	Group     string          `json:"group"              binding:"required"`
	ExpiredAt *time.Time      `json:"expiredAt,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}

type UpdateListDto struct {
	Key       *string         `json:"key,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
	Group     *string         `json:"group,omitempty"`
	ExpiredAt *time.Time      `json:"expiredAt,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}
