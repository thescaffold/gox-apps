package clientlog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ClientLogModule struct{}

func (m *ClientLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ClientLogModule) Exports() []any { return []any{&ClientLogService{}} }

func (m *ClientLogModule) Declarations() []any {
	return []any{&ClientLogService{}, &ClientLogEntity{}}
}
