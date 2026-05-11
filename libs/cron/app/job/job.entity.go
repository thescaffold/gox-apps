package job

import (
	gocron "github.com/awesome-goose/goose/modules/cron"
	"github.com/awesome-goose/goose/modules/sql"
)

type Job = gocron.CronJob

type JobEntity struct {
	*sql.Entity[Job] `inject:""`
}

func (e *JobEntity) OnRegister() {
	e.Hydrate(`"CronJobs"`, []string{"group", "name", "pattern", "status"}, nil, nil, nil, nil, nil, "created_at desc")
}
