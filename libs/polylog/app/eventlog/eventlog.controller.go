package eventlog

import "github.com/thescaffold/gox-packages/libs/core/crud"

type EventLogController struct {
	crud.CrudResource[EventLog, CreateEventLogDto, UpdateEventLogDto]
	entity *EventLogEntity `inject:""`
}

func (c *EventLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[EventLog, CreateEventLogDto, UpdateEventLogDto]{
		Name: "PolylogEventLog", Searchable: []string{"event_id"},
	})
}
