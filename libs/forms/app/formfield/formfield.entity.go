package formfield

import "github.com/awesome-goose/goose/modules/sql"

type FormField struct {
	sql.BaseEntity

	FormId       string  `gorm:"column:form_id;type:varchar(36);not null"      json:"formId"`
	Key          string  `gorm:"column:key;type:varchar(255);not null"         json:"key"`
	Name         string  `gorm:"column:name;type:varchar(255);not null"        json:"name"`
	Type         string  `gorm:"column:type;type:varchar(255);not null"        json:"type"`
	Desc         *string `gorm:"column:desc;type:varchar(255)"                 json:"desc,omitempty"`
	Placeholder  *string `gorm:"column:placeholder;type:varchar(255)"          json:"placeholder,omitempty"`
	Disabled     *bool   `gorm:"column:disabled;type:boolean"                  json:"disabled,omitempty"`
	Required     *bool   `gorm:"column:required;type:boolean"                  json:"required,omitempty"`
	DefaultValue *string `gorm:"column:default_value;type:varchar(255)"        json:"defaultValue,omitempty"`
	Meta         *string `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status       *string `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (FormField) TableName() string { return "FormFormFields" }

type FormFieldEntity struct {
	*sql.Entity[FormField] `inject:""`
}

func (e *FormFieldEntity) OnRegister() {
	e.Hydrate("FormFormFields", []string{"key", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
