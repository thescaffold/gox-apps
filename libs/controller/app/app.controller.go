package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// AppController mirrors ntx-apps/libs/controller/src/app.controller.ts. Three
// endpoints: GET / (hello), POST /register (route ingestion), GET /now (clock).
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(this.appService.getHello()).
// TS calls success() with data only, so the response has no title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Register mirrors TS @Post('register') register(). The payload becomes one
// Http route and one Ws route; each is upserted by (group, service, type, name).
func (c *AppController) Register(dto *RegisterDto) types.Output {
	if err := c.appService.RegisterRoutes(dto); err != nil {
		return response.InternalServerError(
			c.translate(dto, "apps.controller.app.title"),
			err.Error(),
		)
	}
	return response.Success(nil,
		c.translate(dto, "apps.controller.app.title"),
		c.translate(dto, "apps.controller.app.post.register.success"),
		nil)
}

// Now mirrors TS @Get('now') now(): a structured representation of the current
// UTC time using the same field names as the TS payload.
func (c *AppController) Now(dto *NowDto) types.Output {
	now := utils.UTC().Now()
	data := map[string]any{
		"dateTime": now.Format(utils.FormatFullDateTime),
		"time":     now.Format(utils.FormatFullTime),
		"date":     now.Format(utils.FormatFullDate),
		"day":      now.Format("02"),
		"month":    now.Format("01"),
		"year":     now.Format("2006"),
	}
	return response.Success(data,
		c.translate(dto, "apps.controller.app.title"),
		c.translate(dto, "apps.controller.app.get.now.success"),
		nil)
}

// dtoWithCtx is satisfied by every DTO in this package; each carries an
// embedded NTXContext so the request preference can flow through to translate.
type dtoWithCtx interface{ preference() utils.KeyValue }

func (d *RegisterDto) preference() utils.KeyValue { return d.Ctx.Preference }
func (d *NowDto) preference() utils.KeyValue      { return d.Ctx.Preference }
func (d *HelloDto) preference() utils.KeyValue    { return d.Ctx.Preference }

// translate matches TS this.translate(path, null, preference) — pulls the
// request preference off the DTO's embedded NTXContext.
func (c *AppController) translate(dto dtoWithCtx, path string) string {
	if dto == nil {
		return c.lang.Translate(path, nil, nil)
	}
	return c.lang.Translate(path, nil, dto.preference())
}
