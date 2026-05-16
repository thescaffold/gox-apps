package token

import "encoding/json"

// CreateTokenDto mirrors ntx-apps/libs/fuss/src/api/token/dto/create-token.dto.ts.
// userId / clientId / workspaceId are NOT submitted by clients — TS adds them
// in TokenController.morphs.beforeCreate via a payload spread.
type CreateTokenDto struct {
	UserId      string          `json:"userId,omitempty"`
	ClientId    string          `json:"clientId,omitempty"`
	WorkspaceId string          `json:"workspaceId,omitempty"`
	Group       string          `json:"group"           binding:"required"`
	Service     string          `json:"service"         binding:"required"`
	EntityId    string          `json:"entityId"        binding:"required"`
	EntityName  string          `json:"entityName"      binding:"required"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

// UpdateTokenDto mirrors ntx-apps/libs/fuss/src/api/token/dto/update-token.dto.ts.
type UpdateTokenDto struct {
	Group      *string         `json:"group,omitempty"`
	Service    *string         `json:"service,omitempty"`
	EntityId   *string         `json:"entityId,omitempty"`
	EntityName *string         `json:"entityName,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Status     *string         `json:"status,omitempty"`
}
