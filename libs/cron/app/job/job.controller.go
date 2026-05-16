package job

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// JobController mirrors ntx-apps/libs/cron/src/api/job/job.controller.ts.
type JobController struct {
	crud.CrudResource[Job, CreateJobDto, UpdateJobDto]

	entity *JobEntity    `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *JobController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Job, CreateJobDto, UpdateJobDto]{
		// Mirrors TS job.controller.ts:18 name = 'job'.
		Name: "job",
		// Mirrors TS job.controller.ts:19 searchable = ['group','name','pattern'].
		Searchable: []string{"group", "name", "pattern"},
		// Mirrors TS job.controller.ts:21
		// unique = ({ group, name }) => [{ group, name }].
		Unique: func(d *CreateJobDto) []map[string]any {
			return []map[string]any{{"group": d.Group, "name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
