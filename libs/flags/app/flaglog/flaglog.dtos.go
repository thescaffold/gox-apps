package flaglog

type CreateFlagLogDto struct {
	FlagId *string `json:"flagId,omitempty"`
	Limit  *int    `json:"limit,omitempty"`
	Meta   *string `json:"meta,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateFlagLogDto struct {
	Limit  *int    `json:"limit,omitempty"`
	Meta   *string `json:"meta,omitempty"`
	Status *string `json:"status,omitempty"`
}
