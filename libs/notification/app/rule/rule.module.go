package rule

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type RuleModule struct{}

func (m *RuleModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *RuleModule) Exports() []any { return []any{&RuleService{}, &RuleEntity{}} }

func (m *RuleModule) Declarations() []any {
	return []any{&RuleService{}, &RuleEntity{}}
}
