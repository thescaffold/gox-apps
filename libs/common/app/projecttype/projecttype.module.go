package projecttype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ProjectTypeModule struct{}

func (m *ProjectTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ProjectTypeModule) Exports() []any { return []any{&ProjectTypeService{}} }

func (m *ProjectTypeModule) Declarations() []any {
	return []any{&ProjectTypeService{}, &ProjectTypeEntity{}}
}
