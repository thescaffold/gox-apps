package template

type CreateTemplateDto struct {
	Group  string  `json:"group"  binding:"required"`
	Key    string  `json:"key"    binding:"required"`
	Email  *string `json:"email,omitempty"`
	Sms    *string `json:"sms,omitempty"`
	Web    *string `json:"web,omitempty"`
	Mobile *string `json:"mobile,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateTemplateDto struct {
	Email  *string `json:"email,omitempty"`
	Sms    *string `json:"sms,omitempty"`
	Web    *string `json:"web,omitempty"`
	Mobile *string `json:"mobile,omitempty"`
	Status *string `json:"status,omitempty"`
}
