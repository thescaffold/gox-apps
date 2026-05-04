package source

type CreateSourceDto struct {
	UserId      string  `json:"userId"      binding:"required"`
	ClientId    string  `json:"clientId"    binding:"required"`
	WorkspaceId string  `json:"workspaceId" binding:"required"`
	Category    string  `json:"category"    binding:"required"`
	Key         *string `json:"key,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
	TypeId      string  `json:"typeId"      binding:"required"`
	Name        string  `json:"name"        binding:"required"`
	Desc        *string `json:"desc,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UpdateSourceDto struct {
	Name       *string `json:"name,omitempty"`
	Desc       *string `json:"desc,omitempty"`
	Visibility *string `json:"visibility,omitempty"`
	Status     *string `json:"status,omitempty"`
}
