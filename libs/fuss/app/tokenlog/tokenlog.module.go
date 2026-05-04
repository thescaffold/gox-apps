package tokenlog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type TokenLogModule struct{}

func (m *TokenLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *TokenLogModule) Exports() []any { return []any{&TokenLogService{}} }

func (m *TokenLogModule) Declarations() []any {
	return []any{&TokenLogService{}, &TokenLogEntity{}}
}
