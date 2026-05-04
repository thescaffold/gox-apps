package tag

type CreateTagDto struct {
	UserId      *string `json:"userId,omitempty"`
	ClientId    *string `json:"clientId,omitempty"`
	WorkspaceId *string `json:"workspaceId,omitempty"`
	GroupName   string  `json:"groupName"   binding:"required"`
	ServiceName string  `json:"serviceName" binding:"required"`
	EntityName  string  `json:"entityName"  binding:"required"`
	EntityId    string  `json:"entityId"    binding:"required"`
	TypeId      string  `json:"typeId"      binding:"required"`
	Status      *string `json:"status,omitempty"`
}

type UpdateTagDto struct {
	TypeId *string `json:"typeId,omitempty"`
	Status *string `json:"status,omitempty"`
}
