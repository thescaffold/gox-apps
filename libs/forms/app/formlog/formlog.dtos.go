package formlog

import "encoding/json"

// CreateFormLogDto mirrors ntx-apps/libs/forms/src/api/form-log/dto/create-form-log.dto.ts.
type CreateFormLogDto struct {
	FormId      string          `json:"formId"      binding:"required"`
	FormFieldId string          `json:"formFieldId" binding:"required"`
	Key         string          `json:"key"         binding:"required"`
	Value       string          `json:"value"       binding:"required"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

// UpdateFormLogDto mirrors ntx-apps/libs/forms/src/api/form-log/dto/update-form-log.dto.ts.
type UpdateFormLogDto struct {
	FormId      *string         `json:"formId,omitempty"`
	FormFieldId *string         `json:"formFieldId,omitempty"`
	Key         *string         `json:"key,omitempty"`
	Value       *string         `json:"value,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}
