package history

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type HistoryModule struct{}

func (m *HistoryModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *HistoryModule) Exports() []any { return []any{&HistoryService{}} }

func (m *HistoryModule) Declarations() []any {
	return []any{&HistoryService{}, &HistoryEntity{}}
}
