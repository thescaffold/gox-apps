package rate

type CreateRateDto struct {
	Type     string  `json:"type"     binding:"required"`
	PerUnit  *int    `json:"perUnit,omitempty"`
	Currency string  `json:"currency" binding:"required"`
	Status   *string `json:"status,omitempty"`
}

type UpdateRateDto struct {
	Type     *string `json:"type,omitempty"`
	PerUnit  *int    `json:"perUnit,omitempty"`
	Currency *string `json:"currency,omitempty"`
	Status   *string `json:"status,omitempty"`
}
