package environment

import "encoding/json"

// CreateEnvironmentDto mirrors ntx-apps/libs/flags/src/api/environment/dto/create-environment.dto.ts.
// userId / clientId / workspaceId are NOT submitted by clients — TS adds them
// via the (mis-named) hooks.beforeCreate spread, and the gox equivalent does
// the same via Morphs.BeforeCreate.
type CreateEnvironmentDto struct {
	UserId      string          `json:"userId,omitempty"`
	ClientId    string          `json:"clientId,omitempty"`
	WorkspaceId string          `json:"workspaceId,omitempty"`
	TypeId      string          `json:"typeId"           binding:"required"`
	Name        string          `json:"name"             binding:"required"`
	Desc        *string         `json:"desc,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

// UpdateEnvironmentDto mirrors ntx-apps/libs/flags/src/api/environment/dto/update-environment.dto.ts.
type UpdateEnvironmentDto struct {
	TypeId *string         `json:"typeId,omitempty"`
	Name   *string         `json:"name,omitempty"`
	Desc   *string         `json:"desc,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
