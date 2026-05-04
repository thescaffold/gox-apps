package project

import "github.com/awesome-goose/goose/modules/sql"

type Project struct {
	sql.BaseEntity

	UserId      *string `gorm:"column:user_id;type:varchar(36)"       json:"userId,omitempty"`
	ClientId    *string `gorm:"column:client_id;type:varchar(255)"    json:"clientId,omitempty"`
	WorkspaceId *string `gorm:"column:workspace_id;type:varchar(36)"  json:"workspaceId,omitempty"`
	GroupName   string  `gorm:"column:group_name;type:varchar(255);not null"   json:"groupName"`
	ServiceName string  `gorm:"column:service_name;type:varchar(255);not null" json:"serviceName"`
	EntityName  string  `gorm:"column:entity_name;type:varchar(255);not null"  json:"entityName"`
	EntityId    string  `gorm:"column:entity_id;type:varchar(36);not null"     json:"entityId"`
	TypeId      string  `gorm:"column:type_id;type:varchar(36);not null"       json:"typeId"`
	Status      *string `gorm:"column:status;type:varchar(255)"       json:"status,omitempty"`
}

func (Project) TableName() string { return "CommonProjects" }

type ProjectEntity struct {
	*sql.Entity[Project] `inject:""`
}

func (e *ProjectEntity) OnRegister() {
	e.Hydrate("CommonProjects", []string{"group_name", "service_name", "entity_name", "type_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
