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
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "cache", "ok", nil)
}

func (c *AppController) Get(dto *GetDto) types.Output {
	item, err := c.appService.Get(dto.Key, dto.Group)
	if err != nil || item == nil {
		return response.NotFound("cache", "key not found")
	}
	return response.Success(item, "cache", "ok", nil)
}

func (c *AppController) Set(dto *SetDto) types.Output {
	item, err := c.appService.Set(dto.Key, dto.Value, dto.Group, dto.TTL)
	if err != nil {
		return response.InternalServerError("cache", "failed to set key")
	}
	return response.Success(item, "cache", "ok", nil)
}

func (c *AppController) Del(dto *DelDto) types.Output {
	if err := c.appService.Del(dto.Key, dto.Group); err != nil {
		return response.InternalServerError("cache", "failed to delete key")
	}
	return response.Success(nil, "cache", "deleted", nil)
}

func (c *AppController) Flush(dto *FlushDto) types.Output {
	if err := c.appService.Flush(dto.Group); err != nil {
		return response.InternalServerError("cache", "failed to flush group")
	}
	return response.Success(nil, "cache", "flushed", nil)
}
