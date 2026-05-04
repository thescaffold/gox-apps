package client

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ClientModule struct{}

func (m *ClientModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ClientModule) Exports() []any { return []any{&ClientService{}} }

func (m *ClientModule) Declarations() []any {
	return []any{&ClientService{}, &ClientEntity{}}
}
