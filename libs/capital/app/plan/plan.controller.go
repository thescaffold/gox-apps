package plan
import "github.com/thescaffold/gox-packages-core/crud"
type PlanController struct {
	crud.CrudResource[Plan, CreatePlanDto, UpdatePlanDto]
	entity *PlanEntity `inject:""`
}
func (c *PlanController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Plan, CreatePlanDto, UpdatePlanDto]{
		Name: "CapitalPlan", Searchable: []string{"user_id", "workspace_id", "type_id"},
	})
}
