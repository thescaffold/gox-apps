package history

type CreateHistoryDto struct {
	UserId      string  `json:"userId"      binding:"required"`
	ClientId    string  `json:"clientId"    binding:"required"`
	WorkspaceId string  `json:"workspaceId" binding:"required"`
	Type        string  `json:"type"`
	Query       string  `json:"query"       binding:"required"`
	Status      *string `json:"status,omitempty"`
}

type UpdateHistoryDto struct {
	Type   *string `json:"type,omitempty"`
	Query  *string `json:"query,omitempty"`
	Status *string `json:"status,omitempty"`
}
