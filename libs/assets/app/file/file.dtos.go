package file

// CreateFileDto mirrors TS create-file.dto.ts.
type CreateFileDto struct {
	Type   string   `json:"type"   binding:"required"`
	Bucket string   `json:"bucket" binding:"required"`
	Name   string   `json:"name"   binding:"required"`
	Tags   []string `json:"tags,omitempty"`
	Raw    *string  `json:"raw,omitempty"`
	Status *string  `json:"status,omitempty"`
}

// UpdateFileDto mirrors TS update-file.dto.ts.
type UpdateFileDto struct {
	Type   *string  `json:"type,omitempty"`
	Bucket *string  `json:"bucket,omitempty"`
	Name   *string  `json:"name,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Raw    *string  `json:"raw,omitempty"`
	Status *string  `json:"status,omitempty"`
}
