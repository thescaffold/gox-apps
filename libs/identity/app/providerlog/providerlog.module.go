package providerlog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ProviderLogModule struct{}

func (m *ProviderLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ProviderLogModule) Exports() []any { return []any{&ProviderLogService{}} }

func (m *ProviderLogModule) Declarations() []any {
	return []any{&ProviderLogService{}, &ProviderLogEntity{}}
}
