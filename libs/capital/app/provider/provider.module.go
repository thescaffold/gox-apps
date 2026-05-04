package provider

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ProviderModule struct{}

func (m *ProviderModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ProviderModule) Exports() []any { return []any{&ProviderService{}} }

func (m *ProviderModule) Declarations() []any {
	return []any{&ProviderService{}, &ProviderEntity{}}
}
