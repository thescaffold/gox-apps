package template

import "github.com/awesome-goose/goose/modules/sql"

type Template struct {
	sql.BaseEntity

	Group  string  `gorm:"column:group;type:varchar(255);not null" json:"group"`
	Key    string  `gorm:"column:key;type:varchar(255);not null"   json:"key"`
	Email  *string `gorm:"column:email;type:text"                  json:"email,omitempty"`
	Sms    *string `gorm:"column:sms;type:text"                    json:"sms,omitempty"`
	Web    *string `gorm:"column:web;type:text"                    json:"web,omitempty"`
	Mobile *string `gorm:"column:mobile;type:text"                 json:"mobile,omitempty"`
	Status *string `gorm:"column:status;type:varchar(255)"         json:"status,omitempty"`
}

func (Template) TableName() string { return "NotificationTemplates" }

type TemplateEntity struct {
	*sql.Entity[Template] `inject:""`
}

func (e *TemplateEntity) OnRegister() {
	e.Hydrate("NotificationTemplates", []string{"group", "key"}, nil, nil, nil, nil, nil, "created_at desc")
}
