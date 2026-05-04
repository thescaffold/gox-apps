package permission

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type PermissionModule struct{}

func (m *PermissionModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *PermissionModule) Exports() []any { return []any{&PermissionService{}} }

func (m *PermissionModule) Declarations() []any {
	return []any{&PermissionService{}, &PermissionEntity{}}
}
