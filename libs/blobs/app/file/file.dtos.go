package file

import "encoding/json"

type CreateFileDto struct {
	Type     string          `json:"type"             binding:"required"`
	ParentId *string         `json:"parentId,omitempty"`
	Name     string          `json:"name"             binding:"required"`
	Bucket   string          `json:"bucket"           binding:"required"`
	Size     *int64          `json:"size,omitempty"`
	Mime     *string         `json:"mime,omitempty"`
	Tags     json.RawMessage `json:"tags,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

type UpdateFileDto struct {
	Type     *string         `json:"type,omitempty"`
	ParentId *string         `json:"parentId,omitempty"`
	Name     *string         `json:"name,omitempty"`
	Bucket   *string         `json:"bucket,omitempty"`
	Url      *string         `json:"url,omitempty"`
	Size     *int64          `json:"size,omitempty"`
	Mime     *string         `json:"mime,omitempty"`
	Tags     json.RawMessage `json:"tags,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
