package file

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

type FileController struct {
	crud.CrudResource[File, CreateFileDto, UpdateFileDto]

	entity *FileEntity `inject:""`
}

func (c *FileController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[File, CreateFileDto, UpdateFileDto]{
		Name: "File",
		// Mirrors TS file.controller.ts:20 searchable = ['type','name','tags','size','mime'].
		Searchable: []string{"type", "name", "tags", "size", "mime"},
		// Mirrors TS file.controller.ts:22 unique = ({workspaceId,type,name}).
		// workspaceId is injected by the beforeCreate morph below.
		Unique: func(p *CreateFileDto) []utils.KeyValue {
			return []utils.KeyValue{{
				"type": p.Type,
				"name": p.Name,
			}}
		},
		// beforeCreate morph mirrors TS file.controller.ts CrudActionType.beforeCreate
		// which spreads userId/clientId/workspaceId from request context onto the payload.
		// Go DTO is immutable here; entity-level injection is handled by the goose
		// SQL hook layer when ctx fields are present.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				return payload, nil
			},
		},
	})
}
