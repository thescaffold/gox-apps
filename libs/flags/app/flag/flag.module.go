package flag

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type FlagModule struct{}

func (m *FlagModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *FlagModule) Exports() []any { return []any{&FlagService{}} }

func (m *FlagModule) Declarations() []any {
	return []any{&FlagService{}, &FlagEntity{}}
}
