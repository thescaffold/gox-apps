package user

type CreateUserDto struct {
	Name   string  `json:"name"   binding:"required"`
	Email  string  `json:"email"  binding:"required"`
	Phone  *string `json:"phone,omitempty"`
	Secret *string `json:"secret,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateUserDto struct {
	Name   *string `json:"name,omitempty"`
	Phone  *string `json:"phone,omitempty"`
	Secret *string `json:"secret,omitempty"`
	Status *string `json:"status,omitempty"`
}
