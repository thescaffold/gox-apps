package formlog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// FormLogController mirrors ntx-apps/libs/forms/src/api/form-log/form-log.controller.ts.
type FormLogController struct {
	crud.CrudResource[FormLog, CreateFormLogDto, UpdateFormLogDto]
	entity *FormLogEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *FormLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[FormLog, CreateFormLogDto, UpdateFormLogDto]{
		// Mirrors TS form-log.controller.ts:18 name = 'form log'.
		Name: "form log",
		// Mirrors TS form-log.controller.ts:19 searchable = ['key'].
		Searchable: []string{"key"},
		// Mirrors TS form-log.controller.ts:21-24:
		//   unique = ({ formId, formFieldId, key }) => [{ formId, formFieldId }, { formId, key }]
		Unique: func(d *CreateFormLogDto) []map[string]any {
			return []map[string]any{
				{"form_id": d.FormId, "form_field_id": d.FormFieldId},
				{"form_id": d.FormId, "key": d.Key},
			}
		},
	})
	c.SetLang(c.lang)
}
