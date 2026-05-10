package file

import (
	"github.com/thescaffold/gox-packages-core/crud"
)

type FileController struct {
	crud.CrudResource[File, CreateFileDto, UpdateFileDto]

	entity *FileEntity `inject:""`
}

func (c *FileController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[File, CreateFileDto, UpdateFileDto]{
		Name: "File",
		// Mirrors TS file.controller.ts:19 searchable = ['bucket','name','tags']
		Searchable: []string{"bucket", "name", "tags"},
		// Mirrors TS file.controller.ts:21 unique = ({type, bucket, name}) => [{type, bucket, name}]
		Unique: func(payload *CreateFileDto) []map[string]any {
			return []map[string]any{{"type": payload.Type, "bucket": payload.Bucket, "name": payload.Name}}
		},
	})
}
