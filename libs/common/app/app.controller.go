package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "common", "ok", nil)
}

func (c *AppController) GetCurrency(dto *GetCurrencyDto) types.Output {
	rate, err := c.appService.GetCurrency(dto.Currency)
	if err != nil || rate == nil {
		return response.NotFound("common", "currency not found")
	}
	return response.Success(rate, "common", "ok", nil)
}

func (c *AppController) GetLocation(dto *GetLocationDto) types.Output {
	if dto.Ip == "" {
		return response.BadRequest("common", "ip is required")
	}
	loc, err := c.appService.GetLocation(dto.Ip)
	if err != nil {
		return response.InternalServerError("common", "location lookup failed")
	}
	return response.Success(loc, "common", "ok", nil)
}
