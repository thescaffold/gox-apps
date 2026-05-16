package tagtype

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// TagActionDto mirrors ntx-apps/libs/common/src/app.dto.ts TagActionDto.
// Body of POST /tag-type/sync /attach /detach — identifies the target entity
// (group/service/entity/entityId) plus the list of tag-type names to align.
type TagActionDto struct {
	GroupName   string            `json:"groupName"   binding:"required"`
	ServiceName string            `json:"serviceName" binding:"required"`
	EntityName  string            `json:"entityName"  binding:"required"`
	EntityId    string            `json:"entityId"    binding:"required"`
	Names       []string          `json:"names"       binding:"required"`
	Ctx         ntxctx.NTXContext `context:"ntx"`
}

type CreateTagTypeDto struct {
	UserId      *string `json:"userId,omitempty"`
	ClientId    *string `json:"clientId,omitempty"`
	WorkspaceId *string `json:"workspaceId,omitempty"`
	Name        string  `json:"name"   binding:"required"`
	Desc        *string `json:"desc,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UpdateTagTypeDto struct {
	Name   *string `json:"name,omitempty"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}
