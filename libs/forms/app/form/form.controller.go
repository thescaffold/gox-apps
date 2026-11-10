package form

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FormController mirrors ntx-apps/libs/forms/src/api/form/form.controller.ts.
// The TS controller adds a beforeCreate morph that pulls userId/clientId/
// workspaceId off the request context; the gox equivalent populates them on
// the CreateFormDto so the new copyAny carries them onto the Form row's
// NOT NULL columns.
type FormController struct {
	crud.CrudResource[Form, CreateFormDto, UpdateFormDto]
	entity *FormEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *FormController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Form, CreateFormDto, UpdateFormDto]{
		// Mirrors TS form.controller.ts:19 name = 'form'.
		Name: "form",
		// A form is its workspace's (U26).
		WorkspaceScoped: true,
		// Mirrors TS form.controller.ts:20 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS form.controller.ts:22-25:
		//   unique = ({ workspaceId, key, name }) => [{ workspaceId, name }, { workspaceId, key }]
		// Keys are the DB column names because crud.mapToWhere builds raw SQL.
		Unique: func(d *CreateFormDto) []map[string]any {
			return []map[string]any{
				{"workspace_id": d.WorkspaceId, "name": d.Name},
				{"workspace_id": d.WorkspaceId, "key": d.Key},
			}
		},
		// Mirrors TS morphs.beforeCreate: spreads { userId, clientId,
		// workspaceId } from the request context onto the payload.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateFormDto)
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
				// The verified session wins over the body, so the unique check runs in the
				// caller's workspace (U26).
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
