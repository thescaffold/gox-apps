package plantype
import "github.com/thescaffold/gox-packages-core/crud"
type PlanTypeController struct {
	crud.CrudResource[PlanType, CreatePlanTypeDto, UpdatePlanTypeDto]
	entity *PlanTypeEntity `inject:""`
}
func (c *PlanTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[PlanType, CreatePlanTypeDto, UpdatePlanTypeDto]{
		Name: "BridgePlanType", Searchable: []string{"license_id", "key", "name"},
	})
}
