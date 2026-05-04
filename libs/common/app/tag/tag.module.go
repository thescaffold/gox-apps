package tag

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type TagModule struct{}

func (m *TagModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *TagModule) Exports() []any { return []any{&TagService{}} }

func (m *TagModule) Declarations() []any {
	return []any{&TagService{}, &TagEntity{}}
}
