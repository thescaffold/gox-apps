package list

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ListModule struct{}

func (m *ListModule) Imports() []types.Module {
	return []types.Module{
		ROUTES,
		sql.Child(&sql.Config{}),
	}
}

func (m *ListModule) Exports() []any {
	return []any{&ListService{}}
}

func (m *ListModule) Declarations() []any {
	return []any{
		&ListService{},
		&ListEntity{},
	}
}
