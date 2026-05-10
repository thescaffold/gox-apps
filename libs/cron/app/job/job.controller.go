package job

import "github.com/thescaffold/gox-packages/libs/core/crud"

type JobController struct {
	crud.CrudResource[Job, CreateJobDto, UpdateJobDto]

	entity *JobEntity `inject:""`
}

func (c *JobController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Job, CreateJobDto, UpdateJobDto]{
		Name:       "CronJob",
		Searchable: []string{"group", "name", "pattern", "status"},
	})
}
