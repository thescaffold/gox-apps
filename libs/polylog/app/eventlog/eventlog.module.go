package eventlog
import ("github.com/awesome-goose/goose/modules/sql"; "github.com/awesome-goose/goose/types")
type EventLogModule struct{}
func (m *EventLogModule) Imports() []types.Module { return []types.Module{ROUTES, sql.Child(&sql.Config{})} }
func (m *EventLogModule) Exports() []any { return []any{&EventLogService{}, &EventLogEntity{}} }
func (m *EventLogModule) Declarations() []any { return []any{&EventLogService{}, &EventLogEntity{}} }
