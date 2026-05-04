package projecttype

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
