package plan
import ("github.com/awesome-goose/goose/modules/sql"; "github.com/awesome-goose/goose/types")
type PlanModule struct{}
func (m *PlanModule) Imports() []types.Module { return []types.Module{ROUTES, sql.Child(&sql.Config{})} }
func (m *PlanModule) Exports() []any { return []any{&PlanService{}} }
func (m *PlanModule) Declarations() []any { return []any{&PlanService{}, &PlanEntity{}} }
