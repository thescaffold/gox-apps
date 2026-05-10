package event

import "github.com/thescaffold/gox-packages-core/crud"

type EventController struct {
	crud.CrudResource[Event, CreateEventDto, UpdateEventDto]
	entity *EventEntity `inject:""`
}

func (c *EventController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Event, CreateEventDto, UpdateEventDto]{
		Name: "PolylogEvent", Searchable: []string{"user_id", "workspace_id", "entity_name", "reference"},
	})
}
