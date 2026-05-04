package environment

type CreateEnvironmentDto struct {
	UserId      string  `json:"userId"      binding:"required"`
	ClientId    string  `json:"clientId"    binding:"required"`
	WorkspaceId string  `json:"workspaceId" binding:"required"`
	TypeId      string  `json:"typeId"      binding:"required"`
	Name        string  `json:"name"        binding:"required"`
	Desc        *string `json:"desc,omitempty"`
	Meta        *string `json:"meta,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UpdateEnvironmentDto struct {
	TypeId *string `json:"typeId,omitempty"`
	Name   *string `json:"name,omitempty"`
	Desc   *string `json:"desc,omitempty"`
	Meta   *string `json:"meta,omitempty"`
	Status *string `json:"status,omitempty"`
}
