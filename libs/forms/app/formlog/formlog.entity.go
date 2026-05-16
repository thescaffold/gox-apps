package formlog

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type FormLog struct {
	sql.BaseEntity

	FormId      string          `gorm:"column:form_id;type:varchar(36);not null"       json:"formId"`
	FormFieldId string          `gorm:"column:form_field_id;type:varchar(36);not null" json:"formFieldId"`
	Key         string          `gorm:"column:key;type:varchar(255);not null"          json:"key"`
	Value       string          `gorm:"column:value;type:varchar(255);not null"        json:"value"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                         json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"                json:"status,omitempty"`
}

func (FormLog) TableName() string { return "FormFormLogs" }

type FormLogEntity struct {
	*sql.Entity[FormLog] `inject:""`
}

func (e *FormLogEntity) OnRegister() {
	// Searchable mirrors TS form-log.controller.ts:19 searchable = ['key'].
	e.Hydrate("FormFormLogs", []string{"key"}, nil, nil, nil, nil, nil, "created_at desc")
}
