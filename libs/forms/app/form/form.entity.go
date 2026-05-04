package form

import "github.com/awesome-goose/goose/modules/sql"

type Form struct {
	sql.BaseEntity

	UserId      string  `gorm:"column:user_id;type:varchar(36);not null"   json:"userId"`
	ClientId    string  `gorm:"column:client_id;type:varchar(255);not null" json:"clientId"`
	WorkspaceId string  `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	TypeId      string  `gorm:"column:type_id;type:varchar(36);not null"   json:"typeId"`
	Key         string  `gorm:"column:key;type:varchar(255);not null"      json:"key"`
	Name        string  `gorm:"column:name;type:varchar(255);not null"     json:"name"`
	Desc        *string `gorm:"column:desc;type:varchar(255)"              json:"desc,omitempty"`
	Meta        *string `gorm:"column:meta;type:jsonb"                     json:"meta,omitempty"`
	Status      *string `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (Form) TableName() string { return "FormForms" }

type FormEntity struct {
	*sql.Entity[Form] `inject:""`
}

func (e *FormEntity) OnRegister() {
	e.Hydrate("FormForms", []string{"key", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
