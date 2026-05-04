package licensetype

type CreateLicenseTypeDto struct {
	Key      string   `json:"key"      binding:"required"`
	Name     string   `json:"name"     binding:"required"`
	Desc     *string  `json:"desc,omitempty"`
	Detail   *string  `json:"detail,omitempty"`
	Type     *string  `json:"type,omitempty"`
	Currency string   `json:"currency" binding:"required"`
	Daily    *float64 `json:"daily,omitempty"`
	Weekly   *float64 `json:"weekly,omitempty"`
	Monthly  float64  `json:"monthly"  binding:"required"`
	Yearly   *float64 `json:"yearly,omitempty"`
	Meta     *string  `json:"meta,omitempty"`
	Status   *string  `json:"status,omitempty"`
}

type UpdateLicenseTypeDto struct {
	Name     *string  `json:"name,omitempty"`
	Desc     *string  `json:"desc,omitempty"`
	Detail   *string  `json:"detail,omitempty"`
	Type     *string  `json:"type,omitempty"`
	Currency *string  `json:"currency,omitempty"`
	Daily    *float64 `json:"daily,omitempty"`
	Weekly   *float64 `json:"weekly,omitempty"`
	Monthly  *float64 `json:"monthly,omitempty"`
	Yearly   *float64 `json:"yearly,omitempty"`
	Meta     *string  `json:"meta,omitempty"`
	Status   *string  `json:"status,omitempty"`
}
