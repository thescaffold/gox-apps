package file

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// FileController mirrors ntx-apps/libs/blobs/src/api/file/file.controller.ts.
type FileController struct {
	crud.CrudResource[File, CreateFileDto, UpdateFileDto]

	entity *FileEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *FileController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[File, CreateFileDto, UpdateFileDto]{
		// Mirrors TS file.controller.ts:19 name = 'file'.
		Name: "file",
		// A file belongs to a workspace: a caller lists, reads and changes only its own
		// workspace's files, and a new one lands there whatever the body says (U26).
		WorkspaceScoped: true,
		// Mirrors TS file.controller.ts:20 searchable = ['type','name','tags','size','mime'].
		Searchable: []string{"type", "name", "tags", "size", "mime"},
		// Mirrors TS file.controller.ts:22-24:
		//   unique = ({workspaceId, type, name}) => [{workspaceId, type, name}]
		// workspace_id is filled by the beforeCreate morph below.
		Unique: func(p *CreateFileDto) []utils.KeyValue {
			return []utils.KeyValue{{
				"workspace_id": p.WorkspaceId,
				"type":         p.Type,
				"name":         p.Name,
			}}
		},
		// Mirrors TS morphs.beforeCreate: spreads { userId, clientId,
		// workspaceId } from the request context onto the payload.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateFileDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok {
						p.WorkspaceId = id
					}
				}
				// The verified session's ids win over the body, so the unique check below runs in
				// the caller's own workspace (U26).
				if ctx.UserID != "" {
					p.UserId = ctx.UserID
				}
				if ctx.ClientID != "" {
					p.ClientId = ctx.ClientID
				}
				if ctx.WorkspaceID != "" {
					p.WorkspaceId = ctx.WorkspaceID
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
