package rate

type CreateRateDto struct {
	Currency string  `json:"currency" binding:"required"`
	Value    *int    `json:"value,omitempty"`
	Delta    *int    `json:"delta,omitempty"`
	Status   *string `json:"status,omitempty"`
}

type UpdateRateDto struct {
	Value  *int    `json:"value,omitempty"`
	Delta  *int    `json:"delta,omitempty"`
	Status *string `json:"status,omitempty"`
}
