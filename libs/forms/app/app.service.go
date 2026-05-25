package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	formpkg "github.com/thescaffold/gox-apps/libs/forms/app/form"
	formfieldpkg "github.com/thescaffold/gox-apps/libs/forms/app/formfield"
	formlogpkg "github.com/thescaffold/gox-apps/libs/forms/app/formlog"
)

// ErrFormNotFound is returned by GetForm / SaveForm when no row matches the
// supplied :id. The controller translates this to the i18n not-found message.
var ErrFormNotFound = errors.New("form not found")

// FieldNotFoundError is returned by SaveForm when a key in the request body
// does not correspond to a registered form field. The controller translates
// this to `apps.forms.app.error.field-not-found` with `{key}` interpolation.
type FieldNotFoundError struct{ Key string }

func (e *FieldNotFoundError) Error() string { return "field not found: " + e.Key }

// AppService backs the /one/:id endpoints. Mirrors ntx-apps/libs/forms/src/app.controller.ts.
type AppService struct {
	formEntity      *formpkg.FormEntity           `inject:""`
	formFieldEntity *formfieldpkg.FormFieldEntity `inject:""`
	formLogEntity   *formlogpkg.FormLogEntity     `inject:""`
}

// GetHello mirrors TS AppService.getHello().
func (s *AppService) GetHello() string { return "Hello World!" }

// FieldProps mirrors the TS toForm() field.props object exactly.
type FieldProps struct {
	Type        any     `json:"type"`
	Label       string  `json:"label"`
	Description *string `json:"description"`
	Placeholder *string `json:"placeholder"`
	Required    *bool   `json:"required"`
	Options     any     `json:"options"`
	Rows        any     `json:"rows"`
}

// FieldView is one element of the toForm() `fields` array.
type FieldView struct {
	Key          string     `json:"key"`
	Type         string     `json:"type"`
	Props        FieldProps `json:"props"`
	DefaultValue *string    `json:"defaultValue"`
	Validators   any        `json:"validators"`
}

// FormView is the response payload from GET /one/:id, structurally identical
// to the TS toForm() return value (sans the `fn` function which TS drops at
// JSON-stringify time).
type FormView struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Desc        *string        `json:"desc"`
	ButtonTitle string         `json:"buttonTitle"`
	Fields      []FieldView    `json:"fields"`
	Model       map[string]any `json:"model"`
	Disabled    bool           `json:"disabled"`
	Class       string         `json:"class"`
	TitleClass  string         `json:"titleClass"`
	ButtonType  string         `json:"buttonType"`
	ButtonClass string         `json:"buttonClass"`
}

// GetForm returns the toForm() view for a saved form. Mirrors TS getForm().
func (s *AppService) GetForm(id string) (*FormView, error) {
	form, err := s.formEntity.First(`id = ?`, id)
	if err != nil || form == nil {
		return nil, ErrFormNotFound
	}
	fields, err := s.formFieldEntity.Find(0, 0, `form_id = ?`, id)
	if err != nil {
		return nil, err
	}
	return toForm(form, fields), nil
}

// SaveForm writes one FormLog per (key, value) pair after verifying every key
// resolves to a registered form field. Validation aborts before any insert so
// a partial body never produces a partial save, matching TS which builds the
// logs array first then issues a single bulk save.
func (s *AppService) SaveForm(id string, body map[string]any) error {
	form, _ := s.formEntity.First(`id = ?`, id)
	if form == nil {
		return ErrFormNotFound
	}

	// First pass: validate every key has a registered field, materialise rows.
	// Iterate keys in a deterministic (sorted) order so the field-not-found error
	// reports a stable key. (Go maps have no insertion order, so we cannot mirror
	// JS for…in order exactly; sorting removes the prior non-determinism.)
	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := make([]formlogpkg.FormLog, 0, len(body))
	for _, key := range keys {
		raw := body[key]
		field, _ := s.formFieldEntity.First(`form_id = ? AND "key" = ?`, id, key)
		if field == nil {
			return &FieldNotFoundError{Key: key}
		}
		rows = append(rows, formlogpkg.FormLog{
			FormId:      id,
			FormFieldId: field.Id,
			Key:         key,
			Value:       stringify(raw),
		})
	}

	// Second pass: bulk-insert.
	for i := range rows {
		if err := s.formLogEntity.Insert(&rows[i]); err != nil {
			return err
		}
	}
	return nil
}

// toForm mirrors the TS toForm() helper. meta and field.meta are jsonb blobs;
// each is unmarshalled into a map so the field-pick logic (`meta?.foo`)
// translates to a single map lookup with a typed fallback.
func toForm(form *formpkg.Form, fields []formfieldpkg.FormField) *FormView {
	meta := decodeMeta(form.Meta)
	buttonTitle, _ := meta["buttonTitle"].(string)
	if buttonTitle == "" {
		buttonTitle = "Save"
	}

	fieldViews := make([]FieldView, 0, len(fields))
	for _, f := range fields {
		fmeta := decodeMeta(f.Meta)

		props := FieldProps{
			Type:        fmeta["type"],
			Label:       f.Name,
			Description: f.Desc,
			Placeholder: f.Placeholder,
			Required:    f.Required,
			Options:     fmeta["options"],
			Rows:        fmeta["rows"],
		}
		if props.Options == nil {
			props.Options = []any{}
		}
		if props.Rows == nil {
			props.Rows = 3
		}

		validators := fmeta["validators"]
		if validators == nil {
			validators = map[string]any{}
		}

		fieldViews = append(fieldViews, FieldView{
			Key:          f.Key,
			Type:         f.Type,
			Props:        props,
			DefaultValue: f.DefaultValue,
			Validators:   validators,
		})
	}

	return &FormView{
		ID:          form.Id,
		Title:       form.Name,
		Desc:        form.Desc,
		ButtonTitle: buttonTitle,
		Fields:      fieldViews,
		Model:       map[string]any{},
		Disabled:    false,
		Class:       "align-items-start",
		TitleClass:  "nav-link active",
		ButtonType:  "justify-content-end",
		ButtonClass: "text-light",
	}
}

// decodeMeta unmarshalls a jsonb blob into a generic map. Empty / invalid
// content yields a fresh empty map so callers can use a single lookup pattern.
func decodeMeta(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	if out == nil {
		return map[string]any{}
	}
	return out
}

// stringify is gox's analogue of TS storing `body[key]` directly into the
// FormLog.value varchar column. Strings pass through, primitives get their
// natural representation, structured values are JSON-encoded so they
// round-trip cleanly.
func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64, int, int64:
		return fmt.Sprintf("%v", x)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}
