package form

import "encoding/json"

// CreateFormDto mirrors ntx-apps/libs/forms/src/api/form/dto/create-form.dto.ts.
// userId/clientId/workspaceId are NOT submitted by clients — TS adds them in
// FormController.morphs.beforeCreate via a payload spread. They live on the
// DTO here (without binding:"required") so the morph can populate them.
type CreateFormDto struct {
	UserId      string          `json:"userId,omitempty"`
	ClientId    string          `json:"clientId,omitempty"`
	WorkspaceId string          `json:"workspaceId,omitempty"`
	TypeId      string          `json:"typeId,omitempty"`
	Key         string          `json:"key,omitempty"`
	Name        string          `json:"name"               binding:"required"`
	Desc        *string         `json:"desc,omitempty"`
	Meta        json.RawMessage `json:"meta"               binding:"required"`
	Status      *string         `json:"status,omitempty"`
}

// UpdateFormDto mirrors ntx-apps/libs/forms/src/api/form/dto/update-form.dto.ts.
type UpdateFormDto struct {
	TypeId *string         `json:"typeId,omitempty"`
	Key    *string         `json:"key,omitempty"`
	Name   *string         `json:"name,omitempty"`
	Desc   *string         `json:"desc,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
