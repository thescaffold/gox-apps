package formfield

import "encoding/json"

// CreateFormFieldDto mirrors ntx-apps/libs/forms/src/api/form-field/dto/create-form-field.dto.ts.
// TS makes `key` optional and `formId` UUID-required; gox follows the same shape.
type CreateFormFieldDto struct {
	FormId       string          `json:"formId"            binding:"required"`
	Key          string          `json:"key,omitempty"`
	Name         string          `json:"name"              binding:"required"`
	Type         string          `json:"type"              binding:"required"`
	Desc         *string         `json:"desc,omitempty"`
	Placeholder  *string         `json:"placeholder,omitempty"`
	Disabled     *bool           `json:"disabled,omitempty"`
	Required     *bool           `json:"required,omitempty"`
	DefaultValue *string         `json:"defaultValue,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
	Status       *string         `json:"status,omitempty"`
}

// UpdateFormFieldDto mirrors ntx-apps/libs/forms/src/api/form-field/dto/update-form-field.dto.ts.
type UpdateFormFieldDto struct {
	FormId       *string         `json:"formId,omitempty"`
	Key          *string         `json:"key,omitempty"`
	Name         *string         `json:"name,omitempty"`
	Type         *string         `json:"type,omitempty"`
	Desc         *string         `json:"desc,omitempty"`
	Placeholder  *string         `json:"placeholder,omitempty"`
	Disabled     *bool           `json:"disabled,omitempty"`
	Required     *bool           `json:"required,omitempty"`
	DefaultValue *string         `json:"defaultValue,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
	Status       *string         `json:"status,omitempty"`
}
