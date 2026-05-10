package app

import (
	"fmt"

	formpkg "github.com/thescaffold/gox-apps/libs/forms/app/form"
	formfieldpkg "github.com/thescaffold/gox-apps/libs/forms/app/formfield"
	formlogpkg "github.com/thescaffold/gox-apps/libs/forms/app/formlog"
	formtypepkg "github.com/thescaffold/gox-apps/libs/forms/app/formtype"
)

// AppService backs the /one/:id endpoints. Mirrors ntx-apps/libs/forms/src/app.controller.ts.
type AppService struct {
	formEntity      *formpkg.FormEntity           `inject:""`
	formFieldEntity *formfieldpkg.FormFieldEntity `inject:""`
	formLogEntity   *formlogpkg.FormLogEntity     `inject:""`
	formTypeEntity  *formtypepkg.FormTypeEntity   `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

// FormView is the response shape from GET /one/:id, mirroring TS toForm().
// Includes the FormType lookup so clients receive type metadata in one shot
// (matches TS findOne({relations:['type','fields']})).
type FormView struct {
	ID     string                   `json:"id"`
	Key    string                   `json:"key"`
	Name   string                   `json:"name"`
	Desc   *string                  `json:"desc,omitempty"`
	Type   *formtypepkg.FormType    `json:"type,omitempty"`
	Fields []formfieldpkg.FormField `json:"fields"`
}

// GetForm returns the form record + its fields + its type. Mirrors TS getForm().
func (s *AppService) GetForm(id string) (*FormView, error) {
	form, err := s.formEntity.First(`id = ?`, id)
	if err != nil || form == nil {
		return nil, fmt.Errorf("form not found")
	}
	fields, err := s.formFieldEntity.Find(0, 0, `form_id = ?`, id)
	if err != nil {
		return nil, err
	}
	view := &FormView{ID: form.Id, Key: form.Key, Name: form.Name, Desc: form.Desc, Fields: fields}
	// Hydrate the relation: TS uses TypeORM's @ManyToOne with eager join; in Go
	// we resolve the type by id. Empty TypeId leaves Type nil.
	if form.TypeId != "" {
		if ft, _ := s.formTypeEntity.First(`id = ?`, form.TypeId); ft != nil {
			view.Type = ft
		}
	}
	return view, nil
}

// SaveForm validates each (id, key) is a registered field, then writes one
// FormLog row per key/value pair. Mirrors TS saveForm().
func (s *AppService) SaveForm(id string, body map[string]any) error {
	form, _ := s.formEntity.First(`id = ?`, id)
	if form == nil {
		return fmt.Errorf("form not found")
	}
	for key, raw := range body {
		field, _ := s.formFieldEntity.First(`form_id = ? AND "key" = ?`, id, key)
		if field == nil {
			return fmt.Errorf("field %q not found", key)
		}
		value := fmt.Sprintf("%v", raw)
		if err := s.formLogEntity.Insert(&formlogpkg.FormLog{
			FormId: id, FormFieldId: field.Id, Key: key, Value: value,
		}); err != nil {
			return err
		}
	}
	return nil
}
