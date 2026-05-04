package usage

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type UsageModule struct{}

func (m *UsageModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *UsageModule) Exports() []any { return []any{&UsageService{}} }

func (m *UsageModule) Declarations() []any {
	return []any{&UsageService{}, &UsageEntity{}}
}
