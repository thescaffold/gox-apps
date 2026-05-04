package webhook
type CreateWebhookDto struct {
	LicenseId string  `json:"licenseId" binding:"required"`
	Type      *string `json:"type,omitempty"`
	Url       *string `json:"url,omitempty"`
	Status    *string `json:"status,omitempty"`
}
type UpdateWebhookDto struct {
	Url    *string `json:"url,omitempty"`
	Status *string `json:"status,omitempty"`
}
