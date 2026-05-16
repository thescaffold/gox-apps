package user

type CreateUserDto struct {
	Ref    *string `json:"ref,omitempty"`
	Type   *string `json:"type,omitempty"`
	Name   string  `json:"name"   binding:"required"`
	Email  string  `json:"email"  binding:"required"`
	Phone  *string `json:"phone,omitempty"`
	Secret *string `json:"secret,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateUserDto struct {
	Ref    *string `json:"ref,omitempty"`
	Type   *string `json:"type,omitempty"`
	Name   *string `json:"name,omitempty"`
	Phone  *string `json:"phone,omitempty"`
	Secret *string `json:"secret,omitempty"`
	Status *string `json:"status,omitempty"`
}
