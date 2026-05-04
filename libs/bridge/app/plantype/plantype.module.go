package plantype
import ("github.com/awesome-goose/goose/modules/sql"; "github.com/awesome-goose/goose/types")
type PlanTypeModule struct{}
func (m *PlanTypeModule) Imports() []types.Module { return []types.Module{ROUTES, sql.Child(&sql.Config{})} }
func (m *PlanTypeModule) Exports() []any { return []any{&PlanTypeService{}} }
func (m *PlanTypeModule) Declarations() []any { return []any{&PlanTypeService{}, &PlanTypeEntity{}} }
