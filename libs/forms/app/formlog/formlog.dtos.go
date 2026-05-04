package formlog

type CreateFormLogDto struct {
	FormId      string  `json:"formId"      binding:"required"`
	FormFieldId string  `json:"formFieldId" binding:"required"`
	Key         string  `json:"key"         binding:"required"`
	Value       string  `json:"value"       binding:"required"`
	Status      *string `json:"status,omitempty"`
}

type UpdateFormLogDto struct {
	Value  *string `json:"value,omitempty"`
	Status *string `json:"status,omitempty"`
}
