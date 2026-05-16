package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/fuss/src/app.controller.ts. Four
// endpoints under /apps/fuss: GET / (hello), GET /recent (paginated history),
// GET /search (full-text token search), GET /lookup (entity meta by tuple).
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) with no envelope.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Recent mirrors TS @Get('recent') recent(). Note TS uses the same
// `apps.fuss.app.get.search.success` translation key here too — verbatim.
func (c *AppController) Recent(dto *RecentDto) types.Output {
	pref := dto.NTX.Preference
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
		return response.InternalServerError(
			c.lang.Translate("apps.fuss.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Paginated(items, page, perPage, total,
		c.lang.Translate("apps.fuss.app.title", nil, pref),
		c.lang.Translate("apps.fuss.app.get.search.success", nil, pref))
}

// Search mirrors TS @Get('search') search().
func (c *AppController) Search(dto *SearchDto) types.Output {
	pref := dto.NTX.Preference
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
		return response.InternalServerError(
			c.lang.Translate("apps.fuss.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Paginated(items, page, perPage, total,
		c.lang.Translate("apps.fuss.app.title", nil, pref),
		c.lang.Translate("apps.fuss.app.get.search.success", nil, pref))
}

// Lookup mirrors TS @Get('lookup') findOne(). TS returns `token?.meta ?? null`
// — a 200 success with the meta payload (or null when no token row matches).
func (c *AppController) Lookup(dto *LookupDto) types.Output {
	pref := dto.NTX.Preference
	meta, err := c.appService.Lookup(dto.Group, dto.Service, dto.EntityName, dto.EntityId)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.fuss.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(meta,
		c.lang.Translate("apps.fuss.app.title", nil, pref),
		c.lang.Translate("apps.fuss.app.get.lookup.success", nil, pref),
		nil)
}
