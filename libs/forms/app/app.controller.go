package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "forms", "ok", nil)
}

// GetForm handles GET /one/:id. Mirrors TS getForm().
func (c *AppController) GetForm(dto *GetFormDto) types.Output {
	view, err := c.appService.GetForm(dto.Id)
	if err != nil {
		return response.NotFound("forms", err.Error())
	}
	return response.Success(view, "forms", "ok", nil)
}

// SaveForm handles POST /one/:id. Mirrors TS saveForm().
func (c *AppController) SaveForm(dto *SaveFormDto) types.Output {
	if err := c.appService.SaveForm(dto.Id, dto.Body); err != nil {
		return response.BadRequest("forms", err.Error())
	}
	return response.Success(nil, "forms", "ok", nil)
}
