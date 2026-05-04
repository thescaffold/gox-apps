package source

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type SourceModule struct{}

func (m *SourceModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *SourceModule) Exports() []any { return []any{&SourceService{}} }

func (m *SourceModule) Declarations() []any {
	return []any{&SourceService{}, &SourceEntity{}}
}
