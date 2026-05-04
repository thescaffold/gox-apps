package page

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type PageModule struct{}

func (m *PageModule) Imports() []types.Module {
	return []types.Module{
		ROUTES,
		sql.Child(&sql.Config{}),
	}
}

func (m *PageModule) Exports() []any {
	return []any{&PageService{}}
}

func (m *PageModule) Declarations() []any {
	return []any{
		&PageService{},
		&PageEntity{},
	}
}
