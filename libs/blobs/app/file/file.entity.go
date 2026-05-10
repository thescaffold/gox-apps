package file

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

// File mirrors ntx-apps/libs/blobs/src/api/file/entities/file.entity.ts.
// Persisted columns: userId, clientId, workspaceId, type, parentId, name,
// tags, size, mime, status, meta. No bucket/url — those were Go-only drift
// and have been removed for parity.
type File struct {
	sql.BaseEntity

	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	Type        string          `gorm:"column:type;type:varchar(255);not null"        json:"type"`
	ParentId    *string         `gorm:"column:parent_id;type:varchar(36)"             json:"parentId,omitempty"`
	Name        string          `gorm:"column:name;type:varchar(255);not null"        json:"name"`
	Tags        []string        `gorm:"column:tags;type:jsonb;serializer:json"        json:"tags,omitempty"`
	Size        int64           `gorm:"column:size;not null"                          json:"size"`
	Mime        string          `gorm:"column:mime;type:varchar(255);not null"        json:"mime"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
}

func (File) TableName() string { return "BlobsFiles" }

type FileEntity struct {
	*sql.Entity[File] `inject:""`
}

func (e *FileEntity) OnRegister() {
	e.Hydrate(
		"BlobsFiles",
		// Mirrors TS file.controller.ts:20 searchable.
		[]string{"type", "name", "tags", "size", "mime"},
		[]string{},
		nil,
		// unique: (workspaceId, type, name) per TS file.controller.ts:22.
		func(f *File) (any, []any) {
			return map[string]any{
				"workspace_id": f.WorkspaceId,
				"type":         f.Type,
				"name":         f.Name,
			}, nil
		},
		nil,
		nil,
		"created_at desc",
	)
}
