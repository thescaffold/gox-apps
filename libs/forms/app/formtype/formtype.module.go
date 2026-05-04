package formtype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type FormTypeModule struct{}

func (m *FormTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *FormTypeModule) Exports() []any { return []any{&FormTypeService{}} }

func (m *FormTypeModule) Declarations() []any {
	return []any{&FormTypeService{}, &FormTypeEntity{}}
}
