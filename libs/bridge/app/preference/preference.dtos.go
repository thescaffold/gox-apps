package preference
type CreatePreferenceDto struct {
	LicenseId string  `json:"licenseId" binding:"required"`
	Key       string  `json:"key"       binding:"required"`
	Value     *string `json:"value,omitempty"`
	Type      *string `json:"type,omitempty"`
	Status    *string `json:"status,omitempty"`
}
type UpdatePreferenceDto struct {
	Value  *string `json:"value,omitempty"`
	Status *string `json:"status,omitempty"`
}
