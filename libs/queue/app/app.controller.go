package app

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "queue", "ok", nil)
}

func (c *AppController) Push(dto *PushDto) types.Output {
	var config *goqueues.JobConfig
	if dto.Priority != 0 || dto.RetryLimit != 0 || dto.RetryDelay != 0 {
		config = &goqueues.JobConfig{
			Priority:   dto.Priority,
			RetryLimit: dto.RetryLimit,
			RetryDelay: dto.RetryDelay,
		}
	}
	data, err := c.appService.Push(dto.Queue, dto.Job, dto.Data, config)
	if err != nil {
		return response.BadRequest("queue", err.Error())
	}
	if data == nil {
		return response.BadRequest("queue", "invalid request")
	}
	return response.Success(data, "queue", "pushed", nil)
}

func (c *AppController) Pop(dto *PopDto) types.Output {
	data, err := c.appService.Pop(dto.Queue, dto.Job)
	if err != nil {
		return response.BadRequest("queue", err.Error())
	}
	return response.Success(data, "queue", "ok", nil)
}

func (c *AppController) Log(dto *LogDto) types.Output {
	data, err := c.appService.Log(dto.JobId, dto.Status, dto.Output)
	if err != nil {
		return response.BadRequest("queue", err.Error())
	}
	if data == nil {
		return response.BadRequest("queue", "invalid request")
	}
	return response.Success(data, "queue", "logged", nil)
}
