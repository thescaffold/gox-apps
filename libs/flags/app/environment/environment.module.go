package environment

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type EnvironmentModule struct{}

func (m *EnvironmentModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *EnvironmentModule) Exports() []any { return []any{&EnvironmentService{}} }

func (m *EnvironmentModule) Declarations() []any {
	return []any{&EnvironmentService{}, &EnvironmentEntity{}}
}
