package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/audit/src/app.controller.ts. Two HTTP
// endpoints: GET / (getHello) and GET /activities. Entity-event subscriptions
// + heartbeat housekeeping are wired via the Subscriptions map in index.go.
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
	log        types.Log     `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(this.appService.getHello()).
// TS calls success() with data only — no title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Activities mirrors TS @Get('activities') activity(): paginated query over
// AuditLogs where desc is not null, returning the standard envelope with the
// usual pagination meta and translated title/message.
func (c *AppController) Activities(dto *ActivitiesDto) types.Output {
	pref := dto.Ctx.Preference

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
		return response.InternalServerError(
			c.lang.Translate("apps.audit.app.title", nil, pref),
			err.Error(),
		)
	}

	return response.Paginated(logs, page, perPage, total,
		c.lang.Translate("apps.audit.app.title", nil, pref),
		c.lang.Translate("apps.audit.app.get.activities.success", nil, pref))
}
