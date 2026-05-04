package flaglog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type FlagLogModule struct{}

func (m *FlagLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *FlagLogModule) Exports() []any { return []any{&FlagLogService{}} }

func (m *FlagLogModule) Declarations() []any {
	return []any{&FlagLogService{}, &FlagLogEntity{}}
}
