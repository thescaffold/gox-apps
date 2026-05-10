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

// SetNx mirrors TS app.controller.ts setNx — returns 1 on insert, 0 if exists.
func (c *AppController) SetNx(dto *SetDto) types.Output {
	n, err := c.appService.SetNx(dto.Key, dto.Value, dto.Group, dto.TTL)
	if err != nil {
		return response.InternalServerError("cache", "failed to setnx")
	}
	return response.Success(n, "cache", "ok", nil)
}

// GetSet mirrors TS app.controller.ts getSet — returns previous entry.
func (c *AppController) GetSet(dto *SetDto) types.Output {
	prev, err := c.appService.GetSet(dto.Key, dto.Value, dto.Group, dto.TTL)
	if err != nil {
		return response.InternalServerError("cache", "failed to getset")
	}
	return response.Success(prev, "cache", "ok", nil)
}

// TTL mirrors TS app.controller.ts ttl — returns remaining seconds.
//
//	-2: key not found, -1: no expiry, >=0: seconds.
func (c *AppController) TTL(dto *TTLDto) types.Output {
	secs, err := c.appService.TTL(dto.Key, dto.Group)
	if err != nil {
		return response.InternalServerError("cache", "failed to read ttl")
	}
	return response.Success(secs, "cache", "ok", nil)
}

// Incr mirrors TS app.controller.ts incr — atomic integer increment.
func (c *AppController) Incr(dto *IncrDto) types.Output {
	n, err := c.appService.Incr(dto.Key, dto.Group)
	if err != nil {
		return response.InternalServerError("cache", "failed to incr")
	}
	return response.Success(n, "cache", "ok", nil)
}
