package app

type HealthDto struct{}

type GetCurrencyDto struct {
	Currency string `uri:"currency" binding:"required"`
}

type GetLocationDto struct {
	Ip string `form:"ip"`
}
