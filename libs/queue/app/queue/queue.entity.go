package queue

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/modules/sql"
)

type Queue = goqueues.QueueQueue

type QueueEntity struct {
	*sql.Entity[Queue] `inject:""`
}

func (e *QueueEntity) OnRegister() {
	e.Hydrate("queue_queues", []string{"name", "status"}, nil, nil, nil, nil, nil, "created_at desc")
}
