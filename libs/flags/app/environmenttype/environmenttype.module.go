package environmenttype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type EnvironmentTypeModule struct{}

func (m *EnvironmentTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *EnvironmentTypeModule) Exports() []any { return []any{&EnvironmentTypeService{}} }

func (m *EnvironmentTypeModule) Declarations() []any {
	return []any{&EnvironmentTypeService{}, &EnvironmentTypeEntity{}}
}
