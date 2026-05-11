package log

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/modules/sql"
)

type Log = goqueues.QueueLog

type LogEntity struct {
	*sql.Entity[Log] `inject:""`
}

func (e *LogEntity) OnRegister() {
	e.Hydrate(`"QueueLogs"`, []string{"queue_id", "job_id", "status"}, nil, nil, nil, nil, nil, "created_at desc")
}
