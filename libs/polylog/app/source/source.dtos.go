package source

import "encoding/json"

// CreateSourceDto mirrors ntx-apps/libs/polylog/src/api/source/dto/create-source.dto.ts.
// userId/clientId/workspaceId are NOT in the TS DTO — populated by the
// controller morph from request context.
type CreateSourceDto struct {
	UserId      string          `json:"-"`
	ClientId    string          `json:"-"`
	WorkspaceId string          `json:"-"`
	Category    string          `json:"category"    binding:"required"`
	Key         *string         `json:"key,omitempty"`
	Visibility  *string         `json:"visibility,omitempty"`
	TypeId      string          `json:"typeId"      binding:"required"`
	Name        string          `json:"name"        binding:"required"`
	Desc        *string         `json:"desc,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateSourceDto struct {
	Name       *string `json:"name,omitempty"`
	Desc       *string `json:"desc,omitempty"`
	Visibility *string `json:"visibility,omitempty"`
	Status     *string `json:"status,omitempty"`
}
