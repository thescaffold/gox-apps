package file

import (
	"github.com/awesome-goose/goose/modules/sql"
)

type File struct {
	sql.BaseEntity

	Type   string   `gorm:"column:type;type:varchar(255);not null"        json:"type"`
	Bucket string   `gorm:"column:bucket;type:varchar(255);not null"      json:"bucket"`
	Name   string   `gorm:"column:name;type:varchar(255);not null"        json:"name"`
	Tags   []string `gorm:"column:tags;type:jsonb;serializer:json"        json:"tags,omitempty"`
	Raw    *string  `gorm:"column:raw;type:text"                          json:"raw,omitempty"`
	Status *string  `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (File) TableName() string { return "AssetsFiles" }

type FileEntity struct {
	*sql.Entity[File] `inject:""`
}

func (e *FileEntity) OnRegister() {
	e.Hydrate(
		"AssetsFiles",
		// Mirrors TS file.controller.ts:19 searchable = ['bucket','name','tags']
		[]string{"bucket", "name", "tags"},
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
