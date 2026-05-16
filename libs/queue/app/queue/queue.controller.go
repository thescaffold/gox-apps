package queue

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// QueueController mirrors ntx-apps/libs/queue/src/api/queue/queue.controller.ts.
type QueueController struct {
	crud.CrudResource[Queue, CreateQueueDto, UpdateQueueDto]

	entity *QueueEntity  `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *QueueController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Queue, CreateQueueDto, UpdateQueueDto]{
		// Mirrors TS queue.controller.ts:18 name = 'queue'.
		Name: "queue",
		// Mirrors TS queue.controller.ts:19 searchable = ['value'] verbatim
		// (the TS value looks like a placeholder/typo but parity is the goal).
		Searchable: []string{"value"},
		// Mirrors TS queue.controller.ts:21 unique = ({ name }) => [{ name }].
		Unique: func(d *CreateQueueDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
