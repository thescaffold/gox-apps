package file

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FileController mirrors ntx-apps/libs/figs/src/api/file/file.controller.ts.
type FileController struct {
	crud.CrudResource[File, CreateFileDto, UpdateFileDto]

	entity *FileEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *FileController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[File, CreateFileDto, UpdateFileDto]{
		// Mirrors TS file.controller.ts:18 name = 'file'.
		Name: "file",
		// Mirrors TS file.controller.ts:19 searchable = ['name','tags'].
		Searchable: []string{"name", "tags"},
		// Mirrors TS file.controller.ts:21 unique = ({name}) => [{name}].
		Unique: func(d *CreateFileDto) []map[string]any {
			return []map[string]any{{"name": d.Name}}
		},
	})
	c.SetLang(c.lang)
}
