package ratelog

import "time"

type CreateRateLogDto struct {
	RateId string     `json:"rateId"  binding:"required"`
	Value  *int       `json:"value,omitempty"`
	Delta  *int       `json:"delta,omitempty"`
	Date   *time.Time `json:"date,omitempty"`
	Status *string    `json:"status,omitempty"`
}

type UpdateRateLogDto struct {
	Value  *int    `json:"value,omitempty"`
	Delta  *int    `json:"delta,omitempty"`
	Status *string `json:"status,omitempty"`
}
