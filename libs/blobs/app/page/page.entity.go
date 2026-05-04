package page

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type Page struct {
	sql.BaseEntity

	FileId string          `gorm:"column:file_id;type:varchar(36);not null" json:"fileId"`
	Index  int             `gorm:"column:index;not null"                    json:"index"`
	Raw    []byte          `gorm:"column:raw;type:bytea"                    json:"raw,omitempty"`
	Meta   json.RawMessage `gorm:"column:meta;type:jsonb"                   json:"meta,omitempty"`
	Status *string         `gorm:"column:status;type:varchar(255)"          json:"status,omitempty"`
}

func (Page) TableName() string { return "BlobsPages" }

type PageEntity struct {
	*sql.Entity[Page] `inject:""`
}

func (e *PageEntity) OnRegister() {
	e.Hydrate(
		"BlobsPages",
		[]string{"file_id"},
		[]string{},
		nil,
		func(p *Page) (any, []any) {
			return map[string]any{"file_id": p.FileId, "index": p.Index}, nil
		},
		nil,
		nil,
		"created_at desc",
	)
}
