package projecttype

import "github.com/awesome-goose/goose/modules/sql"

type ProjectType struct {
	sql.BaseEntity

	UserId       *string `gorm:"column:user_id;type:varchar(36)"       json:"userId,omitempty"`
	ClientId     *string `gorm:"column:client_id;type:varchar(255)"    json:"clientId,omitempty"`
	WorkspaceId  *string `gorm:"column:workspace_id;type:varchar(36)"  json:"workspaceId,omitempty"`
	Name         string  `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Desc         *string `gorm:"column:desc;type:varchar(255)"         json:"desc,omitempty"`
	ThumbnailUrl *string `gorm:"column:thumbnail_url;type:varchar(255)" json:"thumbnailUrl,omitempty"`
	Status       *string `gorm:"column:status;type:varchar(255)"       json:"status,omitempty"`
}

func (ProjectType) TableName() string { return "CommonProjectTypes" }

type ProjectTypeEntity struct {
	*sql.Entity[ProjectType] `inject:""`
}

func (e *ProjectTypeEntity) OnRegister() {
	e.Hydrate("CommonProjectTypes", []string{"name"}, nil, nil, nil, nil, nil, "created_at desc")
}
