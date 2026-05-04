package file

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type File struct {
	sql.BaseEntity

	Type     string          `gorm:"column:type;type:varchar(255);not null"    json:"type"`
	ParentId *string         `gorm:"column:parent_id;type:varchar(36)"         json:"parentId,omitempty"`
	Name     string          `gorm:"column:name;type:varchar(255);not null"    json:"name"`
	Bucket   string          `gorm:"column:bucket;type:varchar(255);not null"  json:"bucket"`
	Url      *string         `gorm:"column:url;type:varchar(500)"              json:"url,omitempty"`
	Size     *int64          `gorm:"column:size"                              json:"size,omitempty"`
	Mime     *string         `gorm:"column:mime;type:varchar(255)"             json:"mime,omitempty"`
	Tags     json.RawMessage `gorm:"column:tags;type:jsonb"                    json:"tags,omitempty"`
	Meta     json.RawMessage `gorm:"column:meta;type:jsonb"                    json:"meta,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"           json:"status,omitempty"`
}

func (File) TableName() string { return "BlobsFiles" }

type FileEntity struct {
	*sql.Entity[File] `inject:""`
}

func (e *FileEntity) OnRegister() {
	e.Hydrate(
		"BlobsFiles",
		[]string{"name", "type", "bucket"},
		[]string{},
		nil,
		func(f *File) (any, []any) {
			return map[string]any{"name": f.Name, "bucket": f.Bucket}, nil
		},
		nil,
		nil,
		"created_at desc",
	)
}
