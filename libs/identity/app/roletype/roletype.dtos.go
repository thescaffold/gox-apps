package roletype

type CreateRoleTypeDto struct {
	Name   string  `json:"name"   binding:"required"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateRoleTypeDto struct {
	Name   *string `json:"name,omitempty"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}
