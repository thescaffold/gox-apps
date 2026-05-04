package file

import "github.com/thescaffold/gox-packages-core/crud"

type FileController struct {
	crud.CrudResource[File, CreateFileDto, UpdateFileDto]

	entity *FileEntity `inject:""`
}

func (c *FileController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[File, CreateFileDto, UpdateFileDto]{
		Name:       "File",
		Searchable: []string{"name", "type", "bucket"},
	})
}
