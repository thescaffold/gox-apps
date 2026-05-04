package job

import "github.com/thescaffold/gox-packages-core/crud"

type JobController struct {
	crud.CrudResource[Job, CreateJobDto, UpdateJobDto]

	entity *JobEntity `inject:""`
}

func (c *JobController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Job, CreateJobDto, UpdateJobDto]{
		Name:       "QueueJob",
		Searchable: []string{"queue_id", "name", "status"},
	})
}
