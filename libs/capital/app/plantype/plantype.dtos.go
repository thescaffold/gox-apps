package plantype

type CreatePlanTypeDto struct {
	ClientId string  `json:"clientId"  binding:"required"`
	Key      string  `json:"key"       binding:"required"`
	Name     string  `json:"name"      binding:"required"`
	Currency string  `json:"currency"  binding:"required"`
	Monthly  int     `json:"monthly"   binding:"required"`
	Daily    *int    `json:"daily,omitempty"`
	Weekly   *int    `json:"weekly,omitempty"`
	Yearly   *int    `json:"yearly,omitempty"`
	Status   *string `json:"status,omitempty"`
}
type UpdatePlanTypeDto struct {
	Name    *string `json:"name,omitempty"`
	Monthly *int    `json:"monthly,omitempty"`
	Status  *string `json:"status,omitempty"`
}
