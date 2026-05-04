package file

import "github.com/thescaffold/gox-packages-core/crud"

type FileController struct {
	crud.CrudResource[File, CreateFileDto, UpdateFileDto]
	entity *FileEntity `inject:""`
}

func (c *FileController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[File, CreateFileDto, UpdateFileDto]{
		Name:       "FigsFile",
		Searchable: []string{"name", "tags"},
	})
}
