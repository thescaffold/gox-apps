package file

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type File struct {
	sql.BaseEntity
	Name   string          `json:"name"   gorm:"column:name;not null"`
	Input  json.RawMessage `json:"input"  gorm:"type:jsonb;column:input;not null"`
	Output json.RawMessage `json:"output" gorm:"type:jsonb;column:output;not null"`
	Meta   json.RawMessage `json:"meta"   gorm:"type:jsonb;column:meta;not null"`
	Url    *string         `json:"url"    gorm:"column:url"`
	Raw    *string         `json:"raw"    gorm:"type:text;column:raw"`
	Tags   json.RawMessage `json:"tags"   gorm:"type:jsonb;column:tags"`
	Status *string         `json:"status" gorm:"column:status"`
}

func (File) TableName() string { return "FigsFiles" }

type FileEntity struct {
	sql.Entity[File]
}

func (e *FileEntity) OnRegister() {
	e.Hydrate("FigsFiles", []string{"name", "status"}, nil, nil, nil, nil, nil, "created_at desc")
}
