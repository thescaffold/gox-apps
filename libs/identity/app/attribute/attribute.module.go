package attribute

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type AttributeModule struct{}

func (m *AttributeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *AttributeModule) Exports() []any { return []any{&AttributeService{}} }

func (m *AttributeModule) Declarations() []any {
	return []any{&AttributeService{}, &AttributeEntity{}}
}
