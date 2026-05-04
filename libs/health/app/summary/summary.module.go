package summary

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type SummaryModule struct{}

func (m *SummaryModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *SummaryModule) Exports() []any { return []any{&SummaryService{}, &SummaryEntity{}} }

func (m *SummaryModule) Declarations() []any {
	return []any{&SummaryService{}, &SummaryEntity{}}
}
