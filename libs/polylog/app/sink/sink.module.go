package sink

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type SinkModule struct{}

func (m *SinkModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *SinkModule) Exports() []any      { return []any{&SinkService{}, &SinkEntity{}} }
func (m *SinkModule) Declarations() []any { return []any{&SinkService{}, &SinkEntity{}} }
