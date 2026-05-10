package projecttype

import "github.com/thescaffold/gox-packages/libs/core/crud"

type ProjectTypeController struct {
	crud.CrudResource[ProjectType, CreateProjectTypeDto, UpdateProjectTypeDto]
	entity *ProjectTypeEntity `inject:""`
}

func (c *ProjectTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[ProjectType, CreateProjectTypeDto, UpdateProjectTypeDto]{
		Name:       "CommonProjectType",
		Searchable: []string{"name"},
	})
}
