package queue

import "github.com/thescaffold/gox-packages/libs/core/crud"

type QueueController struct {
	crud.CrudResource[Queue, CreateQueueDto, UpdateQueueDto]

	entity *QueueEntity `inject:""`
}

func (c *QueueController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Queue, CreateQueueDto, UpdateQueueDto]{
		Name:       "Queue",
		Searchable: []string{"name", "status"},
	})
}
