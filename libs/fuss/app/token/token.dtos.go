package token

import "encoding/json"

type CreateTokenDto struct {
	UserId      string          `json:"userId"      binding:"required"`
	ClientId    string          `json:"clientId"    binding:"required"`
	WorkspaceId string          `json:"workspaceId" binding:"required"`
	Group       string          `json:"group"       binding:"required"`
	Service     string          `json:"service"     binding:"required"`
	EntityId    string          `json:"entityId"    binding:"required"`
	EntityName  string          `json:"entityName"  binding:"required"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateTokenDto struct {
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`
}
