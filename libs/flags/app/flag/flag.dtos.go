package flag

type CreateFlagDto struct {
	UserId        string  `json:"userId"        binding:"required"`
	ClientId      string  `json:"clientId"      binding:"required"`
	WorkspaceId   string  `json:"workspaceId"   binding:"required"`
	EnvironmentId string  `json:"environmentId" binding:"required"`
	Name          string  `json:"name"          binding:"required"`
	Limit         int     `json:"limit"         binding:"required"`
	Priority      int     `json:"priority"      binding:"required"`
	Level         string  `json:"level"         binding:"required"`
	Meta          string  `json:"meta"          binding:"required"`
	Status        *string `json:"status,omitempty"`
}

type UpdateFlagDto struct {
	Name          *string `json:"name,omitempty"`
	Limit         *int    `json:"limit,omitempty"`
	Priority      *int    `json:"priority,omitempty"`
	Level         *string `json:"level,omitempty"`
	Meta          *string `json:"meta,omitempty"`
	Status        *string `json:"status,omitempty"`
}
