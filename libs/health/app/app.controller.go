package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
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

// Ping handles POST /ping. Mirrors TS AppController.ping().
func (c *AppController) Ping(dto *PingDto) types.Output {
	if err := c.appService.Ping(dto.Name, dto.State, dto.Meta); err != nil {
		return response.NotFound("health", err.Error())
	}
	return response.Success(nil, "health", "ok", nil)
}

// Register handles POST /register. Mirrors TS AppController.register().
func (c *AppController) Register(dto *RegisterServiceDto) types.Output {
	svc, err := c.appService.Register(dto.Name, dto.Desc, dto.Status)
	if err != nil {
		return response.InternalServerError("health", err.Error())
	}
	return response.Success(svc, "health", "ok", nil)
}
