package list

import (
	"encoding/json"
	"time"
)

// CreateListDto mirrors TS create-list.dto.ts. Note that TS exposes neither
// a `group` nor an explicit relationship to one — the cache key alone is the
// uniqueness boundary. The gox entity retains a `group` column (defaulted to
// '' by the migration) for legacy schema parity, but it is not part of the
// client-facing DTO.
type CreateListDto struct {
	Key       string          `json:"key"                binding:"required"`
	Value     json.RawMessage `json:"value,omitempty"`
	ExpiredAt *time.Time      `json:"expiredAt,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}

// UpdateListDto mirrors TS update-list.dto.ts — every field optional.
type UpdateListDto struct {
	Key       *string         `json:"key,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
	ExpiredAt *time.Time      `json:"expiredAt,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}
