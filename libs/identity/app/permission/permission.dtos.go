package permission

type CreatePermissionDto struct {
	UserId      *string `json:"userId,omitempty"`
	ClientId    *string `json:"clientId,omitempty"`
	WorkspaceId *string `json:"workspaceId,omitempty"`
	RoleId      *string `json:"roleId,omitempty"`
	Key         string  `json:"key"      binding:"required"`
	Scope       string  `json:"scope"    binding:"required"`
	Resource    string  `json:"resource" binding:"required"`
	Action      string  `json:"action"   binding:"required"`
	Status      *string `json:"status,omitempty"`
}

type UpdatePermissionDto struct {
	Key      *string `json:"key,omitempty"`
	Scope    *string `json:"scope,omitempty"`
	Resource *string `json:"resource,omitempty"`
	Action   *string `json:"action,omitempty"`
	Status   *string `json:"status,omitempty"`
}
