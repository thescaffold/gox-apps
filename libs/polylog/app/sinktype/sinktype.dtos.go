package sinktype
type CreateSinkTypeDto struct {
	Category   string  `json:"category" binding:"required"`
	Name       string  `json:"name"     binding:"required"`
	Desc       *string `json:"desc,omitempty"`
	Status     *string `json:"status,omitempty"`
}
type UpdateSinkTypeDto struct {
	Name   *string `json:"name,omitempty"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}
