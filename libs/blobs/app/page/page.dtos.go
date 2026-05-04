package page

import "encoding/json"

type CreatePageDto struct {
	FileId string          `json:"fileId"          binding:"required"`
	Index  int             `json:"index"           binding:"required"`
	Raw    []byte          `json:"raw,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}

type UpdatePageDto struct {
	FileId *string         `json:"fileId,omitempty"`
	Index  *int            `json:"index,omitempty"`
	Raw    []byte          `json:"raw,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
