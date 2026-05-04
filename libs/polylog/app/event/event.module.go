package event
import ("github.com/awesome-goose/goose/modules/sql"; "github.com/awesome-goose/goose/types")
type EventModule struct{}
func (m *EventModule) Imports() []types.Module { return []types.Module{ROUTES, sql.Child(&sql.Config{})} }
func (m *EventModule) Exports() []any { return []any{&EventService{}, &EventEntity{}} }
func (m *EventModule) Declarations() []any { return []any{&EventService{}, &EventEntity{}} }
