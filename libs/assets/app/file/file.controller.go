package file

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FileController mirrors ntx-apps/libs/assets/src/api/file/file.controller.ts.
type FileController struct {
	crud.CrudResource[File, CreateFileDto, UpdateFileDto]

	entity *FileEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *FileController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[File, CreateFileDto, UpdateFileDto]{
		// Mirrors TS file.controller.ts:18 name = 'file' — lowercase.
		Name: "file",
		// Mirrors TS file.controller.ts:19 searchable = ['bucket','name','tags'].
		Searchable: []string{"bucket", "name", "tags"},
		// Mirrors TS file.controller.ts:21 unique = ({type, bucket, name}) => [{type, bucket, name}].
		Unique: func(payload *CreateFileDto) []map[string]any {
			return []map[string]any{{"type": payload.Type, "bucket": payload.Bucket, "name": payload.Name}}
		},
	})
	c.SetLang(c.lang)
}
