package project

import "github.com/thescaffold/gox-packages-core/crud"

type ProjectController struct {
	crud.CrudResource[Project, CreateProjectDto, UpdateProjectDto]
	entity *ProjectEntity `inject:""`
}

func (c *ProjectController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Project, CreateProjectDto, UpdateProjectDto]{
		Name:       "CommonProject",
		Searchable: []string{"group_name", "service_name", "entity_name"},
	})
}
