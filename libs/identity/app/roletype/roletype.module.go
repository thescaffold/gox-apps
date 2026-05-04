package roletype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type RoleTypeModule struct{}

func (m *RoleTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *RoleTypeModule) Exports() []any { return []any{&RoleTypeService{}} }

func (m *RoleTypeModule) Declarations() []any {
	return []any{&RoleTypeService{}, &RoleTypeEntity{}}
}
