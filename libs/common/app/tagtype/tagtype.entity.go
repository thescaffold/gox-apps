package tagtype

import "github.com/awesome-goose/goose/modules/sql"

type TagType struct {
	sql.BaseEntity

	UserId      *string `gorm:"column:user_id;type:varchar(36)"      json:"userId,omitempty"`
	ClientId    *string `gorm:"column:client_id;type:varchar(255)"   json:"clientId,omitempty"`
	WorkspaceId *string `gorm:"column:workspace_id;type:varchar(36)" json:"workspaceId,omitempty"`
	Name        string  `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Desc        *string `gorm:"column:desc;type:varchar(255)"        json:"desc,omitempty"`
	Status      *string `gorm:"column:status;type:varchar(255)"      json:"status,omitempty"`
}

func (TagType) TableName() string { return "CommonTagTypes" }

type TagTypeEntity struct {
	*sql.Entity[TagType] `inject:""`
}

func (e *TagTypeEntity) OnRegister() {
	e.Hydrate("CommonTagTypes", []string{"name"}, nil, nil, nil, nil, nil, "created_at desc")
}
