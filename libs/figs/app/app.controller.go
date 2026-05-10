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
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "figs", "ok", nil)
}

func (c *AppController) SaveFile(dto *SaveFileDto) types.Output {
	name := dto.Name
	if name == "" {
		if n, ok := dto.MetaRaw["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		return response.BadRequest("figs", "name is required")
	}

	existing, err := c.appService.FindFileByName(name)
	if err == nil && existing != nil {
		return response.Conflict("figs", "file already exists")
	}

	payload := &Payload{
		Meta:   dto.MetaRaw,
		Input:  dto.InputRaw,
		Output: dto.OutputRaw,
	}
	if payload.Meta == nil {
		payload.Meta = map[string]any{}
	}

	url, err := c.appService.SaveFile(name, payload)
	if err != nil {
		return response.InternalServerError("figs", "processing failed")
	}
	if url == "" {
		return response.BadRequest("figs", "invalid request")
	}

	return response.Success(map[string]any{"url": url}, "figs", "ok", nil)
}
