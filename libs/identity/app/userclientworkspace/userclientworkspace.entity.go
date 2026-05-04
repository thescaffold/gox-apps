package userclientworkspace

import "github.com/awesome-goose/goose/modules/sql"

type UserClientWorkspace struct {
	sql.BaseEntity
	UserId      string  `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string  `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string  `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	RoleId      *string `gorm:"column:role_id;type:varchar(36)"               json:"roleId,omitempty"`
	Status      *string `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}
func (UserClientWorkspace) TableName() string { return "IdentityUserClientWorkspaces" }
type UserClientWorkspaceEntity struct{ *sql.Entity[UserClientWorkspace] `inject:""` }
func (e *UserClientWorkspaceEntity) OnRegister() {
	e.Hydrate("IdentityUserClientWorkspaces", []string{"user_id", "client_id", "workspace_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
