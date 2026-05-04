package permissiontype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type PermissionTypeModule struct{}

func (m *PermissionTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *PermissionTypeModule) Exports() []any { return []any{&PermissionTypeService{}} }

func (m *PermissionTypeModule) Declarations() []any {
	return []any{&PermissionTypeService{}, &PermissionTypeEntity{}}
}
