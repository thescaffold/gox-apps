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
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "audit", "ok", nil)
}

func (c *AppController) Activities(dto *ActivitiesDto) types.Output {
	page := dto.Page
	if page < 1 {
		page = 1
	}
	perPage := dto.PerPage
	if perPage < 1 {
		perPage = 10
	}

	logs, total, err := c.appService.Activities(page, perPage)
	if err != nil {
		return response.InternalServerError("audit", "Failed to fetch activities")
	}

	return response.Paginated(logs, page, perPage, total, "audit", "ok")
}
