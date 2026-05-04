package service

type CreateServiceDto struct {
	Name   string  `json:"name"   binding:"required"`
	Desc   *string `json:"desc,omitempty"`
	Type   *string `json:"type,omitempty"`
	State  *string `json:"state,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateServiceDto struct {
	Name   *string `json:"name,omitempty"`
	Desc   *string `json:"desc,omitempty"`
	Type   *string `json:"type,omitempty"`
	State  *string `json:"state,omitempty"`
	Status *string `json:"status,omitempty"`
}
