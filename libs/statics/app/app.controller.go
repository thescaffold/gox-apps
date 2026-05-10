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
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "statics", "ok", nil)
}

func (c *AppController) FilterByKey(dto *FilterByKeyDto) types.Output {
	items, err := c.appService.FilterByKey(dto.Key, dto.ParentId)
	if err != nil {
		return response.InternalServerError("statics", "Failed to filter by key")
	}
	return response.Success(items, "statics", "ok", nil)
}

func (c *AppController) FindByCode(dto *FindByCodeDto) types.Output {
	item, err := c.appService.FindByCode(dto.Code, dto.ParentId)
	if err != nil {
		return response.NotFound("statics", "Item not found")
	}
	return response.Success(item, "statics", "ok", nil)
}

func (c *AppController) FindByValue(dto *FindByValueDto) types.Output {
	item, err := c.appService.FindByValue(dto.Value, dto.ParentId)
	if err != nil {
		return response.NotFound("statics", "Item not found")
	}
	return response.Success(item, "statics", "ok", nil)
}
