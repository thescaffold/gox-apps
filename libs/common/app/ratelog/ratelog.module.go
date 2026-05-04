package ratelog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type RateLogModule struct{}

func (m *RateLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *RateLogModule) Exports() []any { return []any{&RateLogService{}} }

func (m *RateLogModule) Declarations() []any {
	return []any{&RateLogService{}, &RateLogEntity{}}
}
