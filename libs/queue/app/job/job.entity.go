package job

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/modules/sql"
)

type Job = goqueues.QueueJob

type JobEntity struct {
	*sql.Entity[Job] `inject:""`
}

func (e *JobEntity) OnRegister() {
	e.Hydrate(`"QueueJobs"`, []string{"queue_id", "name", "status"}, nil, nil, nil, nil, nil, "created_at desc")
}
