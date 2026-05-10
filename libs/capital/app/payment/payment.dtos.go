package payment

type CreatePaymentDto struct {
	UserId      string  `json:"userId"      binding:"required"`
	ClientId    string  `json:"clientId"    binding:"required"`
	WorkspaceId string  `json:"workspaceId" binding:"required"`
	PlanId      string  `json:"planId"      binding:"required"`
	Type        string  `json:"type"        binding:"required"`
	Reference   string  `json:"reference"   binding:"required"`
	Amount      *int    `json:"amount,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	Status      *string `json:"status,omitempty"`
}
type UpdatePaymentDto struct {
	Status *string `json:"status,omitempty"`
	Paid   *bool   `json:"paid,omitempty"`
}
