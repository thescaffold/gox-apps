package log

import (
	gocron "github.com/awesome-goose/goose/modules/cron"
	"github.com/awesome-goose/goose/modules/sql"
)

type Log = gocron.CronLog

type LogEntity struct {
	*sql.Entity[Log] `inject:""`
}

func (e *LogEntity) OnRegister() {
	e.Hydrate("cron_logs", []string{"job_id", "status"}, nil, nil, nil, nil, nil, "created_at desc")
}
