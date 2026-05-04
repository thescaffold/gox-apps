package tagtype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type TagTypeModule struct{}

func (m *TagTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *TagTypeModule) Exports() []any { return []any{&TagTypeService{}} }

func (m *TagTypeModule) Declarations() []any {
	return []any{&TagTypeService{}, &TagTypeEntity{}}
}
