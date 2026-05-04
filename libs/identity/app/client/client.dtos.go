package client

type CreateClientDto struct {
	Name   string  `json:"name"   binding:"required"`
	Type   string  `json:"type"   binding:"required"`
	Key    string  `json:"key"    binding:"required"`
	Secret *string `json:"secret,omitempty"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateClientDto struct {
	Name   *string `json:"name,omitempty"`
	Type   *string `json:"type,omitempty"`
	Secret *string `json:"secret,omitempty"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}
