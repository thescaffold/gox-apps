package file

import "encoding/json"

type CreateFileDto struct {
	Type   string          `json:"type"   binding:"required"`
	Bucket string          `json:"bucket" binding:"required"`
	Name   string          `json:"name"   binding:"required"`
	Tags   json.RawMessage `json:"tags,omitempty"`
	Raw    *string         `json:"raw,omitempty"`
	Status *string         `json:"status,omitempty"`
}

type UpdateFileDto struct {
	Type   *string         `json:"type,omitempty"`
	Bucket *string         `json:"bucket,omitempty"`
	Name   *string         `json:"name,omitempty"`
	Tags   json.RawMessage `json:"tags,omitempty"`
	Raw    *string         `json:"raw,omitempty"`
	Status *string         `json:"status,omitempty"`
}
