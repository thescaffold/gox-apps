package form

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type FormModule struct{}

func (m *FormModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *FormModule) Exports() []any { return []any{&FormService{}} }

func (m *FormModule) Declarations() []any {
	return []any{&FormService{}, &FormEntity{}}
}
