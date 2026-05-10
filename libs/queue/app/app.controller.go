package app

import (
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "queue", "ok", nil)
}

func (c *AppController) Push(dto *PushDto) types.Output {
	hasTemporal := dto.StartAt != "" || dto.ExpireAt != "" || dto.Singleton || dto.Frequency != ""
	hasAttempts := dto.Priority != 0 || dto.RetryLimit != 0 || dto.RetryDelay != 0
	var config *goqueues.JobConfig
	if hasTemporal || hasAttempts {
		config = &goqueues.JobConfig{
			Priority:   dto.Priority,
			RetryLimit: dto.RetryLimit,
			RetryDelay: dto.RetryDelay,
			Singleton:  dto.Singleton,
			Frequency:  dto.Frequency,
		}
		if t, err := parseISO(dto.StartAt); err == nil && !t.IsZero() {
			config.StartAt = &t
		}
		if t, err := parseISO(dto.ExpireAt); err == nil && !t.IsZero() {
			config.ExpireAt = &t
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

// parseISO accepts RFC3339 and ISO-8601-ish date strings; returns zero time
// for empty input.
func parseISO(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
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
