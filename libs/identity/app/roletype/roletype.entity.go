package roletype

import "github.com/awesome-goose/goose/modules/sql"

type RoleType struct {
	sql.BaseEntity
	Name   string  `gorm:"column:name;type:varchar(255);not null"  json:"name"`
	Desc   *string `gorm:"column:desc;type:varchar(255)"           json:"desc,omitempty"`
	Status *string `gorm:"column:status;type:varchar(255)"         json:"status,omitempty"`
}

func (RoleType) TableName() string { return "IdentityRoleTypes" }

type RoleTypeEntity struct {
	*sql.Entity[RoleType] `inject:""`
}

func (e *RoleTypeEntity) OnRegister() {
	e.Hydrate("IdentityRoleTypes", []string{"name"}, nil, nil, nil, nil, nil, "created_at desc")
}
