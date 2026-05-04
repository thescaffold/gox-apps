package route

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type RouteModule struct{}

func (m *RouteModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *RouteModule) Exports() []any { return []any{&RouteService{}} }

func (m *RouteModule) Declarations() []any {
	return []any{&RouteService{}, &RouteEntity{}}
}
