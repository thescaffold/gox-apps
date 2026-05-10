package permissiontype

import "github.com/awesome-goose/goose/modules/sql"

type PermissionType struct {
	sql.BaseEntity
	Name     string  `gorm:"column:name;type:varchar(255);not null"     json:"name"`
	Key      string  `gorm:"column:key;type:varchar(255);not null"      json:"key"`
	Desc     *string `gorm:"column:desc;type:varchar(255)"              json:"desc,omitempty"`
	Resource string  `gorm:"column:resource;type:varchar(255);not null" json:"resource"`
	Action   string  `gorm:"column:action;type:varchar(255);not null"   json:"action"`
	Scope    string  `gorm:"column:scope;type:varchar(255);not null"    json:"scope"`
	Status   *string `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (PermissionType) TableName() string { return "IdentityPermissionTypes" }

type PermissionTypeEntity struct {
	*sql.Entity[PermissionType] `inject:""`
}

func (e *PermissionTypeEntity) OnRegister() {
	e.Hydrate("IdentityPermissionTypes", []string{"key", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
