package plantype

type CreatePlanTypeDto struct {
	LicenseId string   `json:"licenseId" binding:"required"`
	Key       string   `json:"key"       binding:"required"`
	Name      string   `json:"name"      binding:"required"`
	Currency  string   `json:"currency"  binding:"required"`
	Monthly   float64  `json:"monthly"   binding:"required"`
	Daily     *float64 `json:"daily,omitempty"`
	Weekly    *float64 `json:"weekly,omitempty"`
	Yearly    *float64 `json:"yearly,omitempty"`
	Status    *string  `json:"status,omitempty"`
}
type UpdatePlanTypeDto struct {
	Name    *string  `json:"name,omitempty"`
	Monthly *float64 `json:"monthly,omitempty"`
	Status  *string  `json:"status,omitempty"`
}
