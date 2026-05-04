package sourcetype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type SourceTypeModule struct{}

func (m *SourceTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *SourceTypeModule) Exports() []any { return []any{&SourceTypeService{}} }

func (m *SourceTypeModule) Declarations() []any {
	return []any{&SourceTypeService{}, &SourceTypeEntity{}}
}
