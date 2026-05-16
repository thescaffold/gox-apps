package projecttype

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// ProjectActionDto mirrors ntx-apps/libs/common/src/app.dto.ts ProjectActionDto.
// Body of POST /project-type/sync /attach /detach.
type ProjectActionDto struct {
	GroupName   string            `json:"groupName"   binding:"required"`
	ServiceName string            `json:"serviceName" binding:"required"`
	EntityName  string            `json:"entityName"  binding:"required"`
	EntityId    string            `json:"entityId"    binding:"required"`
	Names       []string          `json:"names"       binding:"required"`
	Ctx         ntxctx.NTXContext `context:"ntx"`
}

type CreateProjectTypeDto struct {
	UserId       *string `json:"userId,omitempty"`
	ClientId     *string `json:"clientId,omitempty"`
	WorkspaceId  *string `json:"workspaceId,omitempty"`
	Name         string  `json:"name"   binding:"required"`
	Desc         *string `json:"desc,omitempty"`
	ThumbnailUrl *string `json:"thumbnailUrl,omitempty"`
	Status       *string `json:"status,omitempty"`
}

type UpdateProjectTypeDto struct {
	Name         *string `json:"name,omitempty"`
	Desc         *string `json:"desc,omitempty"`
	ThumbnailUrl *string `json:"thumbnailUrl,omitempty"`
	Status       *string `json:"status,omitempty"`
}
