package sinktype
import ("github.com/awesome-goose/goose/modules/sql"; "github.com/awesome-goose/goose/types")
type SinkTypeModule struct{}
func (m *SinkTypeModule) Imports() []types.Module { return []types.Module{ROUTES, sql.Child(&sql.Config{})} }
func (m *SinkTypeModule) Exports() []any { return []any{&SinkTypeService{}, &SinkTypeEntity{}} }
func (m *SinkTypeModule) Declarations() []any { return []any{&SinkTypeService{}, &SinkTypeEntity{}} }
