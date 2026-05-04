package attribute

import "encoding/json"

type CreateAttributeDto struct {
	UserId      string          `json:"userId"                binding:"required"`
	ClientId    *string         `json:"clientId,omitempty"`
	WorkspaceId *string         `json:"workspaceId,omitempty"`
	Key         string          `json:"key"                   binding:"required"`
	Value       *string         `json:"value,omitempty"`
	Type        *string         `json:"type,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateAttributeDto struct {
	Value  *string         `json:"value,omitempty"`
	Type   *string         `json:"type,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
