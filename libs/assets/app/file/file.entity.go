package file

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type File struct {
	sql.BaseEntity

	Type   string          `gorm:"column:type;type:varchar(255);not null"   json:"type"`
	Bucket string          `gorm:"column:bucket;type:varchar(255);not null" json:"bucket"`
	Name   string          `gorm:"column:name;type:varchar(255);not null"   json:"name"`
	Tags   json.RawMessage `gorm:"column:tags;type:jsonb"                   json:"tags,omitempty"`
	Raw    *string         `gorm:"column:raw;type:text"                     json:"raw,omitempty"`
	Status *string         `gorm:"column:status;type:varchar(255)"          json:"status,omitempty"`
}

func (File) TableName() string { return "AssetsFiles" }

type FileEntity struct {
	*sql.Entity[File] `inject:""`
}

func (e *FileEntity) OnRegister() {
	e.Hydrate(
		"AssetsFiles",
		[]string{"name", "bucket", "type"},
		[]string{},
		nil,
		func(f *File) (any, []any) {
			return map[string]any{"type": f.Type, "bucket": f.Bucket, "name": f.Name}, nil
		},
		nil,
		nil,
		"created_at desc",
	)
}
