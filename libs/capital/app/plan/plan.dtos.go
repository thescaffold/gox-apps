package plan
type CreatePlanDto struct {
	UserId      string  `json:"userId"      binding:"required"`
	ClientId    string  `json:"clientId"    binding:"required"`
	WorkspaceId string  `json:"workspaceId" binding:"required"`
	TypeId      string  `json:"typeId"      binding:"required"`
	PeriodType  *string `json:"periodType,omitempty"`
	Status      *string `json:"status,omitempty"`
}
type UpdatePlanDto struct{ Status *string `json:"status,omitempty"` }
