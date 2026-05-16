package log

import "encoding/json"

type CreateLogDto struct {
	Group      string `json:"group"      binding:"required"`
	Service    string `json:"service"    binding:"required"`
	EntityId   string `json:"entityId"   binding:"required"`
	EntityName string `json:"entityName" binding:"required"`
	// Action must be one of LogActionTypeCreate/Read/Update/Delete.
	// Mirrors TS @IsEnum(LogActionType) on create-log.dto.ts.
	Action string          `json:"action"     binding:"required,oneof=create read update delete"`
	Desc   *string         `json:"desc,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
	Status *string         `json:"status,omitempty"`

	// UserId / ClientId / WorkspaceId are NOT submitted by the client — TS
	// adds them in LogController.morphs.beforeCreate via a payload spread
	// (`{ ...payload, userId, clientId, workspaceId }`). They live on the DTO
	// here so the gox beforeCreate morph can populate them before copyAny
	// flows the fields onto the Log entity.
	UserId      string `json:"userId,omitempty"`
	ClientId    string `json:"clientId,omitempty"`
	WorkspaceId string `json:"workspaceId,omitempty"`
}

type UpdateLogDto struct {
	Group      *string         `json:"group,omitempty"`
	Service    *string         `json:"service,omitempty"`
	EntityId   *string         `json:"entityId,omitempty"`
	EntityName *string         `json:"entityName,omitempty"`
	Action     *string         `json:"action,omitempty"`
	Desc       *string         `json:"desc,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Status     *string         `json:"status,omitempty"`
}
