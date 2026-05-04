package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService *AppService `inject:""`
	log        types.Log   `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "health", "ok", nil)
}

func (c *AppController) Check(dto *CheckDto) types.Output {
	entry, err := c.appService.Check(dto.Id)
	if err != nil {
		return response.InternalServerError("health", "check failed")
	}
	return response.Success(entry, "health", "ok", nil)
}
