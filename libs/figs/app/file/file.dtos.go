package file

import "encoding/json"

type CreateFileDto struct {
	Name   string          `json:"name"   binding:"required"`
	Input  json.RawMessage `json:"input"`
	Output json.RawMessage `json:"output"`
	Meta   json.RawMessage `json:"meta"`
	Url    *string         `json:"url,omitempty"`
	Raw    *string         `json:"raw,omitempty"`
	Tags   json.RawMessage `json:"tags,omitempty"`
	Status *string         `json:"status,omitempty"`
}

type UpdateFileDto struct {
	Input  json.RawMessage `json:"input,omitempty"`
	Output json.RawMessage `json:"output,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Url    *string         `json:"url,omitempty"`
	Raw    *string         `json:"raw,omitempty"`
	Tags   json.RawMessage `json:"tags,omitempty"`
	Status *string         `json:"status,omitempty"`
}
