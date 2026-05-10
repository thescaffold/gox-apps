package file

import "encoding/json"

// CreateFileDto mirrors ntx-apps/libs/blobs/src/api/file/dto/create-file.dto.ts.
// userId, clientId, workspaceId are injected by the beforeCreate context morph
// and are not part of the request body.
type CreateFileDto struct {
	Type     string          `json:"type"             binding:"required"`
	ParentId *string         `json:"parentId,omitempty"`
	Name     string          `json:"name"             binding:"required"`
	Tags     []string        `json:"tags,omitempty"`
	Size     int64           `json:"size"             binding:"required"`
	Mime     string          `json:"mime"             binding:"required"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

// UpdateFileDto mirrors ntx-apps/libs/blobs/src/api/file/dto/update-file.dto.ts.
type UpdateFileDto struct {
	Type     *string         `json:"type,omitempty"`
	ParentId *string         `json:"parentId,omitempty"`
	Name     *string         `json:"name,omitempty"`
	Tags     []string        `json:"tags,omitempty"`
	Size     *int64          `json:"size,omitempty"`
	Mime     *string         `json:"mime,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
