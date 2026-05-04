package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct{ appService *AppService `inject:""` }

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "capital", "ok", nil)
}
