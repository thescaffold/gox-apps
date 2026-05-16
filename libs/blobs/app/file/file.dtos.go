package file

import "encoding/json"

// CreateFileDto mirrors ntx-apps/libs/blobs/src/api/file/dto/create-file.dto.ts.
// UserId / ClientId / WorkspaceId are NOT submitted by clients — TS adds them
// in FileController.morphs.beforeCreate via a payload spread, and the gox
// equivalent (see file.controller.go) does the same so copyAny can flow them
// onto the File row's NOT NULL columns.
type CreateFileDto struct {
	UserId      string          `json:"userId,omitempty"`
	ClientId    string          `json:"clientId,omitempty"`
	WorkspaceId string          `json:"workspaceId,omitempty"`
	Type        string          `json:"type"             binding:"required"`
	ParentId    *string         `json:"parentId,omitempty"`
	Name        string          `json:"name"             binding:"required"`
	Tags        []string        `json:"tags,omitempty"`
	Size        int64           `json:"size"             binding:"required"`
	Mime        string          `json:"mime"             binding:"required"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
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
