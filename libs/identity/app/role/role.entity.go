package role

import "github.com/awesome-goose/goose/modules/sql"

type Role struct {
	sql.BaseEntity
	Name        string  `gorm:"column:name;type:varchar(255);not null"      json:"name"`
	ClientId    string  `gorm:"column:client_id;type:varchar(255);not null" json:"clientId"`
	WorkspaceId *string `gorm:"column:workspace_id;type:varchar(36)"        json:"workspaceId,omitempty"`
	Type        string  `gorm:"column:type;type:varchar(255);not null"      json:"type"`
	Status      *string `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}
func (Role) TableName() string { return "IdentityRoles" }
type RoleEntity struct{ *sql.Entity[Role] `inject:""` }
func (e *RoleEntity) OnRegister() {
	e.Hydrate("IdentityRoles", []string{"client_id", "name", "type"}, nil, nil, nil, nil, nil, "created_at desc")
}
