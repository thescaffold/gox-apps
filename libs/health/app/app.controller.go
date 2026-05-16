package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/health/src/app.controller.ts. Three HTTP
// endpoints: GET / (hello), POST /ping (record service state), POST /register
// (upsert a Service row). Sub-controllers under /log /service /summary handle
// the regular CRUD surface.
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
	log        types.Log     `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(this.appService.getHello())
// with no title/message envelope.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Ping mirrors TS @Post('ping') ping(): looks up the service by name, writes
// a Log entry, then updates the service's state column.
//
// Note: TS bails silently with `return;` when the service is unknown, which
// Nest serialises as an empty 200 body — clearly a bug. gox surfaces a
// translated error envelope so clients get a usable signal.
func (c *AppController) Ping(dto *PingDto) types.Output {
	pref := dto.Ctx.Preference

	if err := c.appService.Ping(dto.Name, dto.State, dto.Meta); err != nil {
		return response.BadRequest(
			c.lang.Translate("apps.health.app.title", nil, pref),
			err.Error(),
		)
	}

	return response.Success(nil,
		c.lang.Translate("apps.health.app.title", nil, pref),
		c.lang.Translate("apps.health.app.post.ping.success", nil, pref),
		nil)
}

// Register mirrors TS @Post('register') register(): upserts a Service row by
// name, applying desc + status from the body. TS does NOT pass state through
// — only desc and status reach updateOrCreate — so the gox AppService.Register
// signature is intentionally narrower than RegisterServiceDto.
func (c *AppController) Register(dto *RegisterServiceDto) types.Output {
	pref := dto.Ctx.Preference

	if _, err := c.appService.Register(dto.Name, dto.Desc, dto.Status); err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.health.app.title", nil, pref),
			err.Error(),
		)
	}

	return response.Success(nil,
		c.lang.Translate("apps.health.app.title", nil, pref),
		c.lang.Translate("apps.health.app.post.register.success", nil, pref),
		nil)
}
