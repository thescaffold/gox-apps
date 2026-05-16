package flaglog

import "encoding/json"

// CreateFlagLogDto mirrors ntx-apps/libs/flags/src/api/flag-log/dto/create-flag-log.dto.ts.
// TS marks flagId + limit as required.
type CreateFlagLogDto struct {
	FlagId string          `json:"flagId" binding:"required"`
	Limit  int             `json:"limit"  binding:"required"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}

// UpdateFlagLogDto mirrors ntx-apps/libs/flags/src/api/flag-log/dto/update-flag-log.dto.ts.
type UpdateFlagLogDto struct {
	FlagId *string         `json:"flagId,omitempty"`
	Limit  *int            `json:"limit,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
