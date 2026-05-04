package formfield

type CreateFormFieldDto struct {
	FormId       string  `json:"formId"       binding:"required"`
	Key          string  `json:"key"          binding:"required"`
	Name         string  `json:"name"         binding:"required"`
	Type         string  `json:"type"         binding:"required"`
	Desc         *string `json:"desc,omitempty"`
	Placeholder  *string `json:"placeholder,omitempty"`
	Disabled     *bool   `json:"disabled,omitempty"`
	Required     *bool   `json:"required,omitempty"`
	DefaultValue *string `json:"defaultValue,omitempty"`
	Meta         *string `json:"meta,omitempty"`
	Status       *string `json:"status,omitempty"`
}

type UpdateFormFieldDto struct {
	Key          *string `json:"key,omitempty"`
	Name         *string `json:"name,omitempty"`
	Type         *string `json:"type,omitempty"`
	Desc         *string `json:"desc,omitempty"`
	Placeholder  *string `json:"placeholder,omitempty"`
	Disabled     *bool   `json:"disabled,omitempty"`
	Required     *bool   `json:"required,omitempty"`
	DefaultValue *string `json:"defaultValue,omitempty"`
	Meta         *string `json:"meta,omitempty"`
	Status       *string `json:"status,omitempty"`
}
