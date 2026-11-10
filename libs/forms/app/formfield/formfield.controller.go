package formfield

import (
	"net/http"

	formpkg "github.com/thescaffold/gox-apps/libs/forms/app/form"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FormFieldController mirrors ntx-apps/libs/forms/src/api/form-field/form-field.controller.ts.
type FormFieldController struct {
	crud.CrudResource[FormField, CreateFormFieldDto, UpdateFormFieldDto]
	entity *FormFieldEntity    `inject:""`
	forms  *formpkg.FormEntity `inject:""`
	lang   *i18n.Service       `inject:""`
}

func (c *FormFieldController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FormField, CreateFormFieldDto, UpdateFormFieldDto]{
		// Mirrors TS form-field.controller.ts name = 'form field'.
		Name: "form field",
		// Mirrors TS form-field.controller.ts searchable = ['name', 'desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS form-field.controller.ts unique: (formId,key) and (formId,name).
		Unique: func(d *CreateFormFieldDto) []map[string]any {
			return []map[string]any{
				{"form_id": d.FormId, "key": d.Key},
				{"form_id": d.FormId, "name": d.Name},
			}
		},
		// A field is its form's: the caller reaches only fields of its workspace's forms, and may
		// add or move one only onto a form of its workspace (U26).
		WorkspaceScopeClause: `form_id IN (SELECT id FROM "FormForms" WHERE workspace_id = ?)`,
		HooksCtx: map[string]crud.HookCtxFn{
			crud.BeforeCreate: func(e crud.HookEvent, ctx ntxctx.NTXContext) error {
				if d, ok := e.DTO.(*CreateFormFieldDto); ok {
					return c.ownForm(d.FormId, ctx)
				}
				return nil
			},
			crud.BeforeUpdate: func(e crud.HookEvent, ctx ntxctx.NTXContext) error {
				if d, ok := e.DTO.(*UpdateFormFieldDto); ok && d.FormId != nil {
					return c.ownForm(*d.FormId, ctx)
				}
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}

// ownForm refuses a form that is not in the caller's workspace, as if it did not exist.
func (c *FormFieldController) ownForm(formID string, ctx ntxctx.NTXContext) error {
	if ctx.WorkspaceID != "" && c.forms != nil {
		if f, _ := c.forms.First(`id = ? AND workspace_id = ?`, formID, ctx.WorkspaceID); f != nil {
			return nil
		}
	}
	return &crud.HTTPError{Status: http.StatusNotFound, Message: "form not found"}
}
