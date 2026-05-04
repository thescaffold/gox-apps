package role

type CreateRoleDto struct {
	Name        string  `json:"name"     binding:"required"`
	ClientId    string  `json:"clientId" binding:"required"`
	WorkspaceId *string `json:"workspaceId,omitempty"`
	Type        string  `json:"type"     binding:"required"`
	Status      *string `json:"status,omitempty"`
}

type UpdateRoleDto struct {
	Name        *string `json:"name,omitempty"`
	WorkspaceId *string `json:"workspaceId,omitempty"`
	Type        *string `json:"type,omitempty"`
	Status      *string `json:"status,omitempty"`
}
