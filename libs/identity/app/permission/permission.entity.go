package permission

import "github.com/awesome-goose/goose/modules/sql"

// Permission mirrors ntx-apps/libs/identity/src/api/permission/entities/permission.entity.ts.
// TS pairs a RoleType with a PermissionType via (role_type_id,
// permission_type_id). gox originally used (role_id, key) as the per-row
// permission shape. Both shapes coexist after Phase 4 (shadow-column strategy);
// new code writes to role_type_id+permission_type_id.
type Permission struct {
	sql.BaseEntity
	UserId           *string `gorm:"column:user_id;type:varchar(36)"            json:"userId,omitempty"`
	ClientId         *string `gorm:"column:client_id;type:varchar(255)"         json:"clientId,omitempty"`
	WorkspaceId      *string `gorm:"column:workspace_id;type:varchar(36)"       json:"workspaceId,omitempty"`
	RoleTypeId       *string `gorm:"column:role_type_id;type:varchar(36)"       json:"roleTypeId,omitempty"`
	PermissionTypeId *string `gorm:"column:permission_type_id;type:varchar(36)" json:"permissionTypeId,omitempty"`
	RoleId           *string `gorm:"column:role_id;type:varchar(36)"            json:"roleId,omitempty"`
	Key              string  `gorm:"column:key;type:varchar(255);not null"      json:"key"`
	Scope            string  `gorm:"column:scope;type:varchar(255);not null"    json:"scope"`
	Resource         string  `gorm:"column:resource;type:varchar(255);not null" json:"resource"`
	Action           string  `gorm:"column:action;type:varchar(255);not null"   json:"action"`
	Status           *string `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (Permission) TableName() string { return "IdentityPermissions" }

type PermissionEntity struct {
	*sql.Entity[Permission] `inject:""`
}

func (e *PermissionEntity) OnRegister() {
	e.Hydrate("IdentityPermissions", []string{"user_id", "client_id", "workspace_id", "role_id", "key"}, nil, nil, nil, nil, nil, "created_at desc")
}
