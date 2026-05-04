package ip

type CreateIpDto struct {
	Value   string  `json:"value"   binding:"required"`
	Country string  `json:"country" binding:"required"`
	City    *string `json:"city,omitempty"`
	Status  *string `json:"status,omitempty"`
}

type UpdateIpDto struct {
	Country *string `json:"country,omitempty"`
	City    *string `json:"city,omitempty"`
	Status  *string `json:"status,omitempty"`
}
