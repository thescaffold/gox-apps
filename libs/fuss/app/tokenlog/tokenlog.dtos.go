package tokenlog

type CreateTokenLogDto struct {
	TokenId string  `json:"tokenId" binding:"required"`
	Value   string  `json:"value"   binding:"required"`
	Status  *string `json:"status,omitempty"`
}

type UpdateTokenLogDto struct {
	Value  *string `json:"value,omitempty"`
	Status *string `json:"status,omitempty"`
}
