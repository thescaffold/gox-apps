package role

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type RoleModule struct{}

func (m *RoleModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *RoleModule) Exports() []any { return []any{&RoleService{}} }

func (m *RoleModule) Declarations() []any {
	return []any{&RoleService{}, &RoleEntity{}}
}
