package request

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type RequestModule struct{}

func (m *RequestModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *RequestModule) Exports() []any { return []any{&RequestService{}} }

func (m *RequestModule) Declarations() []any {
	return []any{&RequestService{}, &RequestEntity{}}
}
