package userclientworkspace

type CreateUserClientWorkspaceDto struct {
	UserId      string  `json:"userId"      binding:"required"`
	ClientId    string  `json:"clientId"    binding:"required"`
	WorkspaceId string  `json:"workspaceId" binding:"required"`
	RoleId      *string `json:"roleId,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UpdateUserClientWorkspaceDto struct {
	RoleId *string `json:"roleId,omitempty"`
	Status *string `json:"status,omitempty"`
}
