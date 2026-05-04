package app

import (
	gocron "github.com/awesome-goose/goose/modules/cron"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "cron", "ok", nil)
}

func (c *AppController) Register(dto *RegisterDto) types.Output {
	var config *gocron.CronConfig
	if dto.Priority != 0 || dto.RetryLimit != 0 || dto.RetryDelay != 0 {
		config = &gocron.CronConfig{
			Priority:   dto.Priority,
			RetryLimit: dto.RetryLimit,
			RetryDelay: dto.RetryDelay,
		}
	}
	data, err := c.appService.Register(dto.Group, dto.Name, dto.Pattern, config)
	if err != nil {
		return response.BadRequest("cron", err.Error())
	}
	if data == nil {
		return response.BadRequest("cron", "invalid request")
	}
	return response.Success(data, "cron", "registered", nil)
}

func (c *AppController) Select(dto *SelectDto) types.Output {
	data, err := c.appService.Select(dto.Group, dto.Name)
	if err != nil {
		return response.BadRequest("cron", err.Error())
	}
	return response.Success(data, "cron", "ok", nil)
}

func (c *AppController) Log(dto *LogDto) types.Output {
	data, err := c.appService.Log(dto.JobId, dto.Status, dto.Output)
	if err != nil {
		return response.BadRequest("cron", err.Error())
	}
	if data == nil {
		return response.BadRequest("cron", "invalid request")
	}
	return response.Success(data, "cron", "logged", nil)
}
