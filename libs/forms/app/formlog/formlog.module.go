package formlog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type FormLogModule struct{}

func (m *FormLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *FormLogModule) Exports() []any { return []any{&FormLogService{}} }

func (m *FormLogModule) Declarations() []any {
	return []any{&FormLogService{}, &FormLogEntity{}}
}
