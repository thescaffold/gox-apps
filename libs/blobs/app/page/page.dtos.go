package page

import "encoding/json"

// CreatePageDto mirrors ntx-apps/libs/blobs/src/api/page/dto/create-page.dto.ts:
// fileId optional, index optional, raw required.
type CreatePageDto struct {
	FileId *string         `json:"fileId,omitempty"`
	Index  *int            `json:"index,omitempty"`
	Raw    string          `json:"raw"             binding:"required"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}

type UpdatePageDto struct {
	FileId *string         `json:"fileId,omitempty"`
	Index  *int            `json:"index,omitempty"`
	Raw    *string         `json:"raw,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
