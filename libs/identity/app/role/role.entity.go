package role

import "github.com/awesome-goose/goose/modules/sql"

// Role mirrors ntx-apps/libs/identity/src/api/role/entities/role.entity.ts.
// TS pairs (entity_name, entity_id, role_type_id) for the role join; gox
// originally used (name, type, workspace_id). Phase H adds the TS shadow
// columns alongside the legacy ones so both reads remain valid.
type Role struct {
	sql.BaseEntity
	Name        string  `gorm:"column:name;type:varchar(255);not null"      json:"name"`
	ClientId    string  `gorm:"column:client_id;type:varchar(255);not null" json:"clientId"`
	WorkspaceId *string `gorm:"column:workspace_id;type:varchar(36)"        json:"workspaceId,omitempty"`
	Type        string  `gorm:"column:type;type:varchar(255);not null"      json:"type"`
	EntityName  *string `gorm:"column:entity_name;type:varchar(255)"        json:"entityName,omitempty"`
	EntityId    *string `gorm:"column:entity_id;type:varchar(36)"           json:"entityId,omitempty"`
	RoleTypeId  *string `gorm:"column:role_type_id;type:varchar(36)"        json:"roleTypeId,omitempty"`
	Status      *string `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}

func (Role) TableName() string { return "IdentityRoles" }

type RoleEntity struct {
	*sql.Entity[Role] `inject:""`
}

func (e *RoleEntity) OnRegister() {
	e.Hydrate("IdentityRoles", []string{"client_id", "name", "type"}, nil, nil, nil, nil, nil, "created_at desc")
}
