package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "fuss", "ok", nil)
}

func (c *AppController) Recent(dto *RecentDto) types.Output {
	page := dto.Page
	if page < 1 {
		page = 1
	}
	perPage := dto.PerPage
	if perPage < 1 {
		perPage = 20
	}
	items, total, err := c.appService.Recent(
		dto.NTX.UserID, dto.NTX.ClientID, dto.NTX.WorkspaceID,
		page, perPage,
	)
	if err != nil {
		return response.InternalServerError("fuss", "failed to fetch history")
	}
	return response.Paginated(items, page, perPage, total, "fuss", "ok")
}

func (c *AppController) Search(dto *SearchDto) types.Output {
	page := dto.Page
	if page < 1 {
		page = 1
	}
	perPage := dto.PerPage
	if perPage < 1 {
		perPage = 20
	}
	histType := dto.Type
	if histType == "" {
		histType = "global"
	}
	items, total, err := c.appService.Search(
		dto.Query, histType,
		dto.NTX.UserID, dto.NTX.ClientID, dto.NTX.WorkspaceID,
		page, perPage,
	)
	if err != nil {
		return response.InternalServerError("fuss", "search failed")
	}
	return response.Paginated(items, page, perPage, total, "fuss", "ok")
}

func (c *AppController) Lookup(dto *LookupDto) types.Output {
	meta, err := c.appService.Lookup(dto.Group, dto.Service, dto.EntityName, dto.EntityId)
	if err != nil {
		return response.InternalServerError("fuss", "lookup failed")
	}
	return response.Success(meta, "fuss", "ok", nil)
}
