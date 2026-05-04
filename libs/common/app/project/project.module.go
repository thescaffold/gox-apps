package project

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ProjectModule struct{}

func (m *ProjectModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ProjectModule) Exports() []any { return []any{&ProjectService{}} }

func (m *ProjectModule) Declarations() []any {
	return []any{&ProjectService{}, &ProjectEntity{}}
}
