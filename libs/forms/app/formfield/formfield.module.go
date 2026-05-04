package formfield

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type FormFieldModule struct{}

func (m *FormFieldModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *FormFieldModule) Exports() []any { return []any{&FormFieldService{}} }

func (m *FormFieldModule) Declarations() []any {
	return []any{&FormFieldService{}, &FormFieldEntity{}}
}
