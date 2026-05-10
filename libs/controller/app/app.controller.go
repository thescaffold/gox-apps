package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "controller", "ok", nil)
}

func (c *AppController) GetRoutes(dto *GetRoutesDto) types.Output {
	length := dto.Length
	if length <= 0 {
		length = 100
	}
	data, err := c.appService.GetRoutes(length)
	if err != nil {
		return response.InternalServerError("controller", "failed to fetch routes")
	}
	return response.Success(data, "controller", "ok", nil)
}

func (c *AppController) GetRoute(dto *GetRouteDto) types.Output {
	upstream, err := c.appService.GetRoute(dto.Path)
	if err != nil {
		return response.InternalServerError("controller", "route lookup failed")
	}
	if upstream == "" {
		return response.NotFound("controller", "route not found")
	}
	return response.Success(map[string]any{"upstream": upstream}, "controller", "ok", nil)
}
