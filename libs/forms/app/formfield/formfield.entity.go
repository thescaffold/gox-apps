package formfield

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

// FormField mirrors ntx-apps/libs/forms/src/api/form-field/entities/form-field.entity.ts.
type FormField struct {
	sql.BaseEntity

	FormId       string  `gorm:"column:form_id;type:varchar(36);not null" json:"formId"`
	Key          string  `gorm:"column:key;type:varchar(255);not null"    json:"key"`
	Name         string  `gorm:"column:name;type:varchar(255);not null"   json:"name"`
	Type         string  `gorm:"column:type;type:varchar(255);not null"   json:"type"`
	Desc         *string `gorm:"column:desc;type:varchar(255)"            json:"desc,omitempty"`
	Placeholder  *string `gorm:"column:placeholder;type:varchar(255)"     json:"placeholder,omitempty"`
	Disabled     *bool   `gorm:"column:disabled;type:boolean"             json:"disabled,omitempty"`
	Required     *bool   `gorm:"column:required;type:boolean"             json:"required,omitempty"`
	DefaultValue *string `gorm:"column:default_value;type:varchar(255)"   json:"defaultValue,omitempty"`
	// Meta is a jsonb blob — toForm() reads `field.meta?.type/options/rows/...`
	// so it must round-trip as raw JSON rather than a quoted string.
	Meta   json.RawMessage `gorm:"column:meta;type:jsonb"          json:"meta,omitempty"`
	Status *string         `gorm:"column:status;type:varchar(255)" json:"status,omitempty"`
}

func (FormField) TableName() string { return "FormFormFields" }

type FormFieldEntity struct {
	*sql.Entity[FormField] `inject:""`
}

func (e *FormFieldEntity) OnRegister() {
	// Searchable mirrors TS form-field.controller.ts:19 searchable = ['name','desc'].
	e.Hydrate("FormFormFields", []string{"name", "desc"}, nil, nil, nil, nil, nil, "created_at desc")
}
